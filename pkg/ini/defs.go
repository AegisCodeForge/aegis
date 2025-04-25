package ini

import "maps"

type INISection struct {
	Name string
	SubName string
	Value map[string]string
}

type INI map[string]map[string]INISection

func (cfg INI) InsertConfigBatch(name string, subname string, value map[string]string) {
	v, ok := cfg[name]
	if !ok {
		cfg[name] = make(map[string]INISection, 0)
		v = cfg[name]
	}
	vv, ok := v[subname]
	if !ok {
		v[subname] = INISection{Name: name, SubName: subname, Value: value}
		vv = v[subname]
	} else {
		maps.Copy(vv.Value, value)
	}
}

func (cfg INI) InsertValue(name string, subname string, key string, value string) {
	v, ok := cfg[name]
	if !ok {
		cfg[name] = make(map[string]INISection, 0)
		v = cfg[name]
	}
	vv, ok := v[subname]
	if !ok {
		v[subname] = INISection{
			Name: name,
			SubName: subname,
			Value: make(map[string]string, 0),
		}
		vv = v[subname]
	}
	vv.Value[key] = value
}

func (cfg INI) GetValue(name string, subname string, key string) (string, bool) {
	v, ok := cfg[name]
	if !ok {
		cfg[name] = make(map[string]INISection, 0)
		v = cfg[name]
	}
	vv, ok := v[subname]
	if !ok {
		v[subname] = INISection{
			Name: name,
			SubName: subname,
			Value: make(map[string]string, 0),
		}
		vv = v[subname]
	}
	val, ok := vv.Value[key]
	return val, ok
}

