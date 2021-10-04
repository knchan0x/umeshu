package view

import "html/template"

// FuncMap is a wrapper of map[string]interface{}, it use to pass
// FuncMap to HTML template render.
type FuncMap map[string]interface{}

// View ViewManager manages html templates.
type ViewManager struct {
	*template.Template
	funcMap template.FuncMap
}

var Manager *ViewManager

// NewManager creates and returns new view manager instance.
// ViewManager can only be created once.
func NewManager(pattern string, funcMap FuncMap) *ViewManager {
	if Manager != nil {
		return Manager
	}

	castMap := template.FuncMap(funcMap)
	Manager = &ViewManager{
		funcMap: castMap,
	}
	Manager.Template = template.Must(template.New("").Funcs(castMap).ParseGlob(pattern))
	return Manager
}

// SetFuncMap sets FuncMap for HTML render.
// It pass funcMap to *template.Template.FuncMap().
func (m *ViewManager) SetFuncMap(funcMap FuncMap) {
	m.funcMap = template.FuncMap(funcMap)
}
