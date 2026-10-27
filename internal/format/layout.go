package format

import (
	"maps"
	"slices"
	"strings"

	cml "github.com/grespyrad/CMLGo"
)

// layoutPlan связывает переносы с AST, не восстанавливая текст из значений модели.
type layoutPlan struct {
	continuation []int
	breaks       map[int]int
	indexes      map[int]int
	tokens       []cml.Lexeme
	aliases      map[string]string
}

func planLayout(model *cml.Node, tokens []cml.Lexeme, dialect cml.Dialect) layoutPlan {
	p := layoutPlan{
		continuation: make([]int, len(tokens)),
		breaks:       map[int]int{},
		indexes:      map[int]int{},
		tokens:       tokens,
		aliases:      cml.RussianAliases(),
	}
	maps.Copy(p.aliases, dialect.Aliases)

	for i, token := range tokens {
		p.indexes[token.Position.Offset] = i
	}

	p.visit(model)
	p.attachComments()
	p.wrapCommentedParameters()

	return p
}

func (p *layoutPlan) keyword(index int) string {
	text := p.tokens[index].Text
	if canonical, ok := p.aliases[text]; ok {
		return canonical
	}

	return text
}

func (p *layoutPlan) boundary(index, lines int) {
	p.breaks[index] = max(p.breaks[index], lines)
}

func (p *layoutPlan) visit(node *cml.Node) {
	if node == nil {
		return
	}

	for field, values := range node.Fields {
		for _, value := range values {
			switch {
			case value.Node != nil:
				p.childBoundary(node, field, value.Node)
				p.visit(value.Node)
			case field == "doc":
				p.boundary(p.indexes[value.Position.Offset]+1, 1)
			case member(node.Kind, propertyKinds):
				p.property(value)
			default:
			}
		}
	}
}

func (p *layoutPlan) childBoundary(parent *cml.Node, field string, child *cml.Node) {
	if field == "features" && parent.Kind == "UseCase" {
		p.feature(parent, child)

		return
	}

	if member(field, inlineFields) {
		return
	}

	index := p.indexes[child.Position.Offset]

	if field == "consequences" {
		p.propertyAt(child.Position.Offset)
	}

	lines := 1

	if member(child.Kind, blockKinds) || parent.Kind == "ContextMappingModel" && child.Kind != "Import" {
		lines = 2
	}

	p.boundary(index, lines)
}

func (p *layoutPlan) property(value cml.ASTValue) {
	index := p.indexes[value.Position.Offset]
	if index == 0 {
		return
	}

	if value.Text == "true" && p.tokens[index].Kind != "string" {
		if propertyKeyword(p.keyword(index)) {
			p.boundary(index, 1)
		}

		return
	}

	p.propertyAt(value.Position.Offset)
}

func (p *layoutPlan) propertyAt(offset int) {
	index := p.previousCode(p.indexes[offset])
	if index >= 0 && p.tokens[index].Text == "=" {
		index = p.previousCode(index)
	}

	if index >= 0 && propertyKeyword(p.keyword(index)) {
		p.boundary(index, 1)
	}
}

func (p *layoutPlan) previousCode(index int) int {
	for index--; index >= 0; index-- {
		if p.tokens[index].Kind != commentKind {
			return index
		}
	}

	return -1
}

// attachComments переносит границу объявления перед его самостоятельными комментариями.
func (p *layoutPlan) attachComments() {
	for i := len(p.tokens) - 1; i >= 0; i-- {
		lines := p.breaks[i]
		if lines == 0 {
			continue
		}

		start := i
		for start > 0 && p.tokens[start-1].Kind == commentKind && !p.trailing(start-1) {
			start--
		}

		if start != i {
			p.boundary(start, lines)

			p.breaks[i] = 1
		}
	}
}

func (p *layoutPlan) trailing(index int) bool {
	if index == 0 {
		return false
	}

	previous := p.tokens[index-1]
	endLine := previous.Position.Line + strings.Count(previous.Text, "\n")

	return previous.Kind != commentKind && previous.Text != "{" && endLine == p.tokens[index].Position.Line
}

func member(value, list string) bool { return slices.Contains(strings.Fields(list), value) }

const (
	inlineFields = "returnType parameterType parameters stateTransition target to action " +
		"eventProduction delegateHolder oppositeHolder subscribe publish"
	blockKinds = "ContextMap BoundedContext Domain Subdomain Aggregate SculptorModule Application " +
		"Entity ValueObject DomainEvent CommandEvent Service Resource Consumer Repository Trait BasicType " +
		"DataTransferObject Enum Flow Coordination UseCase UserStory Stakeholders StakeholderGroup " +
		"Stakeholder ValueRegister ValueCluster Value ValueElicitation ValueEpic " +
		"ValueNarrative ValueWeigthing Invariant " +
		"Partnership SharedKernel UpstreamDownstreamRelationship CustomerSupplierRelationship"
	propertyKinds = "ContextMap BoundedContext Domain Subdomain Aggregate SculptorModule Entity ValueObject " +
		"DomainEvent CommandEvent Service Resource Consumer Repository Trait BasicType DataTransferObject Enum " +
		"UseCase UserStory Stakeholder ValueCluster Value ValueElicitation Invariant " +
		"UpstreamDownstreamRelationship CustomerSupplierRelationship Partnership SharedKernel"
	propertyKeywords = "contains type state domainVisionStatement responsibilities implementationTechnology " +
		"knowledgeLevel businessModel evolution owner useCases userStories features userRequirements " +
		"likelihoodForChange structuralVolatility contentVolatility availabilityCriticality consistencyCriticality " +
		"storageSimilarity securityCriticality securityZone securityAccessGroup aggregateRoot package gap nogap " +
		"hint webservice scaffold path optimisticLocking auditable cache immutable persistent belongsTo " +
		"databaseTable discriminatorValue discriminatorColumn discriminatorType discriminatorLength inheritanceType " +
		"validate external basePackage aggregateLifecycle ordinal actor secondaryActors interactions benefit scope " +
		"level influence interest description core isCore demonstrator relatedValue opposingValue " +
		"priority impact consequences " +
		"exposedAggregates downstreamRights rule rationale testedBy implementedBy"
)

func (p *layoutPlan) feature(parent, child *cml.Node) {
	index := p.indexes[child.Position.Offset]
	if parent.Fields["features"][0].Position.Offset == child.Position.Offset {
		p.propertyAt(child.Position.Offset)

		return
	}

	p.boundary(index, 1)

	p.continuation[index] = 1
}

func propertyKeyword(keyword string) bool {
	return member(keyword, propertyKeywords) || keyword == "As a" || keyword == "As an" || keyword == "so that"
}

const commentKind = "comment"
