package domain

type Word struct {
	Word         string
	LangKey      string
	Field        string
	FieldDesc    string
	Description  string
	Definition   string
	Locality     string
	Synonyms     []*Word
	Correlated   []*Word
	Translations []*Word
}

type SimpleWord struct {
	Word    string `json:"w"`
	LangTag string `json:"t"`
}

type SimpleWordsPair struct {
	First  []*SimpleWord
	Second []*SimpleWord
}

type Alphabetic []*SimpleWord

func (a Alphabetic) Len() int { return len(a) }

func (a Alphabetic) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a Alphabetic) Less(i, j int) bool {
	return a[i].Word < a[j].Word
}
