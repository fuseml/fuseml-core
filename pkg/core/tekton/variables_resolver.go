package tekton

import (
	"fmt"
	"regexp"
	"strings"
)

type variablesResolver struct {
	references map[string]string
}

func newVariablesResolver() *variablesResolver {
	r := &variablesResolver{}
	r.references = make(map[string]string)
	return r
}

func (r *variablesResolver) clone() (res *variablesResolver) {
	res = newVariablesResolver()
	for key, value := range r.references {
		res.references[key] = value
	}
	return res
}

func (r *variablesResolver) resolve(value string) string {
	regex := regexp.MustCompile(`{([^{{{}}}]*)}`)
	matches := regex.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		if v, ok := r.references[value]; ok {
			return v
		}
	}
	for _, v := range matches {
		toResolve := strings.TrimSpace(v[1])
		resolvedTo, existsReference := r.references[toResolve]
		// if value begins with 'steps.' it could mean that the value is derived from a task output, in that case transalate
		// it to the tekton way of getting a task output "$(tasks.TASK.results.VARIABLE)", or that the value is referencing a step
		// input/output that is resolvable here. For example, referencing an image step output "{{ steps.builder.outputs.mlflow-env }}",
		// it starts with steps but a reference to it exists and resolving it returns:
		// "registry.fuseml-registry/mlflow-builder/{{ inputs.mlflow-codeset.name }}:{{ inputs.mlflow-codeset.version }}" which also
		// have variables to also be resolved.
		if strings.HasPrefix(toResolve, "steps.") {
			// if there is already a reference to 'value' on 'sources' use it,
			// note that the reference might also be parametrized, so we also need to resolve it
			if existsReference {
				resolvedTo = r.resolve(resolvedTo)
			} else {
				replacer := strings.NewReplacer("steps", "tasks", "outputs", "results")
				resolvedTo = fmt.Sprintf("$(%s)", replacer.Replace(toResolve))
			}
		}
		value = strings.ReplaceAll(value, fmt.Sprintf("{{ %s }}", toResolve), resolvedTo)
	}
	return value
}

func (r *variablesResolver) addReference(ref, value string) {
	r.references[ref] = value
}
