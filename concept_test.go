package smartlogic

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestConceptMarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		concept       Concept
		expectedJSON  string
		expectedError bool
	}{
		{
			name: "minimal concept",
			concept: Concept{
				PrefLabel:    "Test Person",
				Type:         TypePerson,
				SchemaObject: ConceptSchemaPerson,
			},
			expectedJSON: `{"skosxl:prefLabel":[{"skosxl:literalForm":[{"@value":"Test Person","@language":"en"}],"@type":["skosxl:Label"]}],"@type":["skos:Concept","http://www.ft.com/ontology/person/Person"],"skos:topConceptOf":{"@id":"http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"}}`,
			/*
				{
				  "skosxl:prefLabel": [
					{
					  "skosxl:literalForm": [
						{
						  "@value": "Test Person",
						  "@language": "en"
						}
					  ],
					  "@type": [
						"skosxl:Label"
					  ]
					}
				  ],
				  "@type": [
					"skos:Concept",
					"http://www.ft.com/ontology/person/Person"
				  ],
				  "skos:topConceptOf": {
					"@id": "http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"
				  }
				}
			*/
			expectedError: false,
		},
		{
			name: "deprecated concept",
			concept: Concept{
				PrefLabel:    "Test Deprecated Person",
				Type:         TypePerson,
				SchemaObject: ConceptSchemaPerson,
				IsDeprecated: true,
			},
			expectedJSON: `{"skosxl:prefLabel":[{"skosxl:literalForm":[{"@value":"Test Deprecated Person","@language":"en"}],"@type":["skosxl:Label"]}],"@type":["skos:Concept","http://www.ft.com/ontology/person/Person"],"skos:topConceptOf":{"@id":"http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"},"http://www.ft.com/ontology/isDeprecated":[true]}`,
			/*
				{
				  "skosxl:prefLabel": [
					{
					  "skosxl:literalForm": [
						{
						  "@value": "Test Deprecated Person",
						  "@language": "en"
						}
					  ],
					  "@type": [
						"skosxl:Label"
					  ]
					}
				  ],
				  "@type": [
					"skos:Concept",
					"http://www.ft.com/ontology/person/Person"
				  ],
				  "skos:topConceptOf": {
					"@id": "http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"
				  },
				  "http://www.ft.com/ontology/isDeprecated": [
					true
				  ]
				}
			*/
			expectedError: false,
		},
		{
			name: "full concept",
			concept: Concept{
				PrefLabel:          "Test Person All Fields",
				AltLabels:          []string{"Short Name"},
				Description:        "New test person",
				Type:               TypePerson,
				SchemaObject:       ConceptSchemaPerson,
				TMEIdentifier:      "TnN0ZWluX09OX0ZvcnR1bmVDb21wYW55X0FBUEw=-T04=",
				FactsetIdentifier:  "000C7F-E",
				WikidataIdentifier: "http://www.wikidata.org/entity/Q312",
			},
			expectedJSON: `{"skosxl:prefLabel":[{"skosxl:literalForm":[{"@value":"Test Person All Fields","@language":"en"}],"@type":["skosxl:Label"]}],"skosxl:altLabel":[{"skosxl:literalForm":[{"@value":"Short Name","@language":"en"}],"@type":["skosxl:Label"]}],"http://www.ft.com/ontology/description":[{"@value":"New test person","@language":"en"}],"@type":["skos:Concept","http://www.ft.com/ontology/person/Person"],"skos:topConceptOf":{"@id":"http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"},"http://www.ft.com/ontology/TMEIdentifier":[{"@value":"TnN0ZWluX09OX0ZvcnR1bmVDb21wYW55X0FBUEw=-T04="}],"http://www.ft.com/ontology/factsetIdentifier":[{"@value":"000C7F-E"}],"http://www.ft.com/ontology/wikidataIdentifier":[{"@value":"http://www.wikidata.org/entity/Q312","@type":"xsd:anyURI"}]}`,
			/*
				{
				  "skosxl:prefLabel": [
				    {
				      "skosxl:literalForm": [
				        {
				          "@value": "Test Person All Fields",
				          "@language": "en"
				        }
				      ],
				      "@type": [
				        "skosxl:Label"
				      ]
				    }
				  ],
				  "skosxl:altLabel": [
				    {
				      "skosxl:literalForm": [
				        {
				          "@value": "Short Name",
				          "@language": "en"
				        }
				      ],
				      "@type": [
				        "skosxl:Label"
				      ]
				    }
				  ],
				  "http://www.ft.com/ontology/description": [
				    {
				      "@value": "New test person",
				      "@language": "en"
				    }
				  ],
				  "@type": [
				    "skos:Concept",
				    "http://www.ft.com/ontology/person/Person"
				  ],
				  "skos:topConceptOf": {
				    "@id": "http://www.ft.com/thing/ConceptScheme/8e564c83-669c-48d5-a208-81fb88a32802"
				  },
				  "http://www.ft.com/ontology/TMEIdentifier": [
				    {
				      "@value": "TnN0ZWluX09OX0ZvcnR1bmVDb21wYW55X0FBUEw=-T04="
				    }
				  ],
				  "http://www.ft.com/ontology/factsetIdentifier": [
				    {
				      "@value": "000C7F-E"
				    }
				  ],
				  "http://www.ft.com/ontology/wikidataIdentifier": [
				    {
				      "@value": "http://www.wikidata.org/entity/Q312",
				      "@type": "xsd:anyURI"
				    }
				  ]
				}
			*/
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonRes, err := json.Marshal(test.concept)
			if err != nil && !test.expectedError {
				t.Errorf("unexpected error marshalling concept: %v", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("expected error marshalling concept")
			}
			if !bytes.Equal(jsonRes, []byte(test.expectedJSON)) {
				t.Errorf("unexpected json returned, got %v, want %v", string(jsonRes), test.expectedJSON)
			}
		})
	}
}
