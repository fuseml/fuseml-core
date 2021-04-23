package main

import (
	"fmt"
	"regexp"
	"strings"
)

type variablesResolver struct {
	references map[string]string
}

func NewVariablesResolver() *variablesResolver {
	r := &variablesResolver{}
	r.references = make(map[string]string)
	return r
}

func (r *variablesResolver) resolve(value string) string {
	regex := regexp.MustCompile(`{([^{{{}}}]*)}`)
	matches := regex.FindAllStringSubmatch(value, -1)
	for _, v := range matches {
		toResolve := strings.TrimSpace(v[1])
		resolvedTo, existsReference := r.references[toResolve]
		// if value begins with 'steps.' it could mean that the value is derived from a
		// task output, in that case transalate it to the tekton way of getting a task
		// output "$(tasks.TASK.results.VARIABLE), or that the value is referencing a step
		// input/output that is resolvable here."
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
