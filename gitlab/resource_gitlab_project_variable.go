package gitlab

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabProjectVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectVariableCreate,
		Read:   resourceGitlabProjectVariableRead,
		Update: resourceGitlabProjectVariableUpdate,
		Delete: resourceGitlabProjectVariableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"key": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: StringIsGitlabVariableName,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"variable_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "env_var",
				ValidateFunc: StringIsGitlabVariableType,
			},
			"protected": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"masked": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"environment_scope": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceGitlabProjectVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	variableType := stringToVariableType(d.Get("variable_type").(string))
	protected := d.Get("protected").(bool)
	masked := d.Get("masked").(bool)
	environmentScope := d.Get("environment_scope").(string)

	options := gitlab.CreateProjectVariableOptions{
		Key:              &key,
		Value:            &value,
		VariableType:     variableType,
		Protected:        &protected,
		Masked:           &masked,
		EnvironmentScope: &environmentScope,
	}
	log.Printf("[DEBUG] create gitlab project variable %s/%s", project, key)

	_, _, err := client.ProjectVariables.CreateVariable(project, &options)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(&project, &key))

	return resourceGitlabProjectVariableRead(d, meta)
}

func resourceGitlabProjectVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project, key, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] read gitlab project variable %s/%s", project, key)
	vars, _, err := client.ProjectVariables.ListVariables(project, &gitlab.ListProjectVariablesOptions{})
	if err != nil {
		return err
	}

	var v *gitlab.ProjectVariable
	for _, vv := range vars {
		if vv.Key == key && vv.EnvironmentScope == d.Get("environment_scope").(string) {
			v = vv
			break
		}
	}

	if v == nil {
		return fmt.Errorf("Variable %s/%s, scope:%s Not Found", project, key, d.Get("environment_scope").(string))
	}

	d.Set("key", v.Key)
	d.Set("value", v.Value)
	d.Set("variable_type", v.VariableType)
	d.Set("project", project)
	d.Set("protected", v.Protected)
	d.Set("masked", v.Masked)
	d.Set("environment_scope", v.EnvironmentScope)
	return nil
}

func resourceGitlabProjectVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	variableType := stringToVariableType(d.Get("variable_type").(string))
	protected := d.Get("protected").(bool)
	masked := d.Get("masked").(bool)
	environmentScope := d.Get("environment_scope").(string)

	options := &gitlab.UpdateProjectVariableOptions{
		Value:            &value,
		VariableType:     variableType,
		Protected:        &protected,
		Masked:           &masked,
		EnvironmentScope: &environmentScope,
	}
	log.Printf("[DEBUG] update gitlab project variable %s/%s", project, key)

	_, _, err := client.ProjectVariables.UpdateVariable(project, key, options)
	if err != nil {
		return err
	}

	return resourceGitlabProjectVariableRead(d, meta)
}

func resourceGitlabProjectVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	key := d.Get("key").(string)
	log.Printf("[DEBUG] Delete gitlab project variable %s/%s", project, key)

	_, err := client.ProjectVariables.RemoveVariable(project, key)
	return err
}
