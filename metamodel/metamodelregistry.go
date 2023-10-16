package metamodel

type Columns struct {
	Name      string
	FieldType string
}

type MetaData struct {
	SecondsToLive int64
	Bytes         int64
	DoPersist     bool
	UpdateQuery   string
	GetQuery      string
	CreateQuery   string
	Columns       []Columns
}

var metadata = make(map[string]MetaData)

func RegisterModel(name string, model MetaData) {
	metadata[name] = model
}

func GetModel(name string) (*MetaData, bool) {
	result, ok := metadata[name]
	if ok {
		return &result, true
	} else {
		return nil, false
	}
}
