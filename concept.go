package smartlogic

import "encoding/json"

const (
	// Concept types defined and available in the FT Ontology, required when creating new concept.
	TypeTopic        = "http://www.ft.com/ontology/Topic"
	TypePerson       = "http://www.ft.com/ontology/person/Person"
	TypeOrganisation = "http://www.ft.com/ontology/organisation/Organisation"
	TypeLocation     = "http://www.ft.com/ontology/Location"
	TypeGenre        = "http://www.ft.com/ontology/Genre"
	TypeBrand        = "http://www.ft.com/ontology/product/Brand"

	// Concept schemas defined and available in the FT Production Model, required when creating new concept.
	// Please not that concepts schemas are not the same as topic type and they are defined per Model.
	// If you are about to create a concept under no schema, please be aware that this concept won't be visible in the
	// Smartlogic UI.
	ConceptSchemaTopic        = "http://www.ft.com/ontology/scheme/Topics"
	ConceptSchemaPerson       = "http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"
	ConceptSchemaOrganisation = "http://www.ft.com/ontology/scheme/Organisations"
	ConceptSchemaLocation     = "http://www.ft.com/thing/ConceptScheme/ae342e72-e8a3-41e4-aaf4-180506750948"
	ConceptSchemaGenre        = "http://www.ft.com/ontology/scheme/9639ccc7-e58e-403f-b80a-88e915a98804"
	ConceptSchemaBrand        = "http://www.ft.com/ontology/scheme/Brands"
	ConceptSchemaAuthor       = "http://www.ft.com/ontology/scheme/Authors"
)

type Concept struct {
	PrefLabel   string
	AltLabels   []string
	Description string

	Type         string
	SchemaObject string

	TMEIdentifier      string
	FactsetIdentifier  string
	WikidataIdentifier string

	IsDeprecated bool
}

func (c Concept) MarshalJSON() ([]byte, error) {
	input := inputConcept{
		PrefLabel: []conceptLabel{{
			LiteralForm: []wordValue{
				{
					Value:    c.PrefLabel,
					Language: "en",
				},
			},
			Type: []string{"skosxl:Label"},
		}},
		Type: []string{"skos:Concept", c.Type},
		TopConceptOf: conceptID{
			ID: c.SchemaObject,
		},
	}
	if c.Description != "" {
		input.Description = []wordValue{
			{
				Value:    c.Description,
				Language: "en",
			},
		}
	}
	for _, al := range c.AltLabels {
		input.AltLabels = append(input.AltLabels, conceptLabel{

			LiteralForm: []wordValue{
				{
					Value:    al,
					Language: "en",
				},
			},
			Type: []string{"skosxl:Label"},
		})
	}
	if c.TMEIdentifier != "" {
		input.TMEIdentifier = []conceptValue{
			{
				Value: c.TMEIdentifier,
			},
		}
	}
	if c.FactsetIdentifier != "" {
		input.FactsetIdentifier = []conceptValue{
			{
				Value: c.FactsetIdentifier,
			},
		}
	}
	if c.WikidataIdentifier != "" {
		input.WikidataIdentifier = []uriValue{
			{
				Value: c.WikidataIdentifier,
				Type:  "xsd:anyURI",
			},
		}
	}
	// we want to set isDeprecated property in the json-ld representation of the concept only when its value is true
	if c.IsDeprecated {
		input.IsDeprecated = []bool{c.IsDeprecated}
	}
	return json.Marshal(input)
}

// inputConcept is helper struct matching the required input format for creating new concept in the Smartlogic API
type inputConcept struct {
	PrefLabel   []conceptLabel `json:"skosxl:prefLabel,omitempty"`
	AltLabels   []conceptLabel `json:"skosxl:altLabel,omitempty"`
	Description []wordValue    `json:"http://www.ft.com/ontology/description,omitempty"`

	Type         []string  `json:"@type,omitempty"`
	TopConceptOf conceptID `json:"skos:topConceptOf,omitempty"`

	TMEIdentifier      []conceptValue `json:"http://www.ft.com/ontology/TMEIdentifier,omitempty"`
	FactsetIdentifier  []conceptValue `json:"http://www.ft.com/ontology/factsetIdentifier,omitempty"`
	WikidataIdentifier []uriValue     `json:"http://www.ft.com/ontology/wikidataIdentifier,omitempty"`

	IsDeprecated []bool `json:"http://www.ft.com/ontology/isDeprecated,omitempty"`
}

type conceptValue struct {
	Value string `json:"@value"`
}

type wordValue struct {
	Value    string `json:"@value"`
	Language string `json:"@language,omitempty"`
}

type uriValue struct {
	Value string `json:"@value"`
	Type  string `json:"@type"`
}

type conceptID struct {
	ID string `json:"@id"`
}

type conceptLabel struct {
	LiteralForm []wordValue `json:"skosxl:literalForm,omitempty"`
	Type        []string    `json:"@type,omitempty"`
}
