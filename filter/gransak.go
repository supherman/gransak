package filter

import (
	"reflect"
	"strings"
)

func NewGransak() *Gransak {
	return &Gransak{
		separator:   "_",
		placeholder: "{{.}}",
		valueholder: "{{v}}",
	}
}

type Gransak struct {
	toEvaluate      []string
	evaluatedTokens []string
	template        string
	separator       string
	placeholder     string
	valueholder     string
	pos             int
	param           *gransakParam
}

func (this *Gransak) ToSql(input string, param interface{}) string {
	this.reset()

	this.tokenize(input)

	this.param = newGransakParam(param, reflect.TypeOf(param).String())

	for this.pos = 0; this.pos < len(this.toEvaluate); this.pos++ {
		token := this.toEvaluate[this.pos]

		if node, isCandidate := isCandidateToOperator(token); isCandidate {
			if foundNode, found := this.find(node, this.pos); found {

				foundNode.Apply(this)

			} else {

				this.evaluated(token)

			}
		} else {

			this.evaluated(token)

		}
	}

	this.replaceValue()

	return strings.Trim(this.template, " ")
}

func (this *Gransak) reset() {
	this.toEvaluate = []string{}
	this.evaluatedTokens = []string{}
	this.template = ""
	this.param = nil
}

func (this *Gransak) tokenize(input string) {
	this.toEvaluate = strings.Split(input, this.separator)
}

func (this *Gransak) find(nodeParam *Node, pos int) (*Node, bool) {
	if pos >= len(this.toEvaluate) {
		return nil, false
	}

	next := this.toEvaluate[pos]

	if nodeParam.Name != next {
		return nil, false
	}

	if len(nodeParam.Nodes) > 0 {
		for _, node := range nodeParam.Nodes {
			if foundNode, found := this.find(node, pos+1); found {

				return foundNode, true

			}
		}

		//none of its children nodes matched, check if is itself an operator
		if nodeParam.IsOperator == true {
			this.pos = pos
			return nodeParam, true
		}

	} else {
		this.pos = pos
		return nodeParam, true
	}

	return nil, false
}

func (this *Gransak) appendField() string {
	field := this.getLastField()
	this.template += field + " " + this.placeholder + " "
	return field
}

func (this *Gransak) getLastField() string {
	field := strings.Join(this.evaluatedTokens, this.separator)
	this.evaluatedTokens = []string{}
	return field
}

func (this *Gransak) evaluated(token string) {
	this.evaluatedTokens = append(this.evaluatedTokens, token)
}

func (this *Gransak) replace(replace, replaceFor string) {
	this.template = strings.Replace(
		this.template,
		replace,
		replaceFor,
		-1,
	)
}

func (this *Gransak) replacePlaceholder(replaceFor string) {
	this.replace(this.placeholder, replaceFor)
}

func (this *Gransak) replaceValueHolder(replaceFor string) {
	this.replace(this.valueholder, replaceFor)
}

func (this *Gransak) replaceValue() {
	if len(this.param.parts) == 0 {
		this.replaceValueHolder(this.param.StrRepresentation)
	} else {
		for _, value := range this.param.parts {
			this.template = strings.Replace(this.template, this.valueholder, value, 1)
		}
	}
}

func (this *Gransak) getCorrectSqlFormat(value string) string {
	if this.param.kind == "string" {
		return "'" + value + "'"
	}
	return value
}

func isCandidateToOperator(item string) (*Node, bool) {
	for _, node := range Tree.Nodes {
		if node.Name == item {
			return node, true
		}
	}
	return &Node{}, false
}
