package spdl

import "encoding/json"

/* Eratta:
the make with capacity arguments helps prevent the superfluous order
of the input [] from being shuffled.  This makes testing easier
*/

func (res *FlagList) UnmarshalJSON(data []byte) error {
	var strs []string
	err := json.Unmarshal(data, &strs)
	if err != nil {
		return err
	}
	fl := NewFlagList(len(strs))
	for _, str := range strs {
		flag, err := FromString(str)
		if err != nil {
			return err
		}
		fl.Add(flag)
	}
	*res = fl
	return nil
}

func (fl FlagList) MarshalJSON() ([]byte, error) {
	if len(fl.ordered) == 0 {
		return []byte("[]"), nil
	}
	strs := make([]string, 0, len(fl.ordered))
	for _, f := range fl.ordered {
		strs = append(strs, f.String())
	}
	return json.Marshal(strs)
}

func (res *FlatFlagList) UnmarshalJSON(data []byte) error {
	var strs []string
	err := json.Unmarshal(data, &strs)
	if err != nil {
		return err
	}
	fl := NewFlatFlagList(len(strs))
	for _, str := range strs {
		flag, err := FlatFromString(str)
		if err != nil {
			return err
		}
		fl.Add(flag)
	}
	*res = fl
	return nil
}

func (fl FlatFlagList) MarshalJSON() ([]byte, error) {
	if len(fl.ordered) == 0 {
		return []byte("[]"), nil
	}
	strs := make([]string, 0, len(fl.ordered))
	for _, f := range fl.ordered {
		strs = append(strs, f.String())
	}
	return json.Marshal(strs)
}

func (fs *FlagExpr) UnmarshalJSON(data []byte) (err error) {
	var str string

	err = json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	*fs, err = fromString(str)
	return err
}

func (fs FlagExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(fs.String())
}

func (d *Dep) UnmarshalJSON(data []byte) (err error) {
	var str string

	err = json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	*d, err = ParseDep(str)
	return err
}

func (d Dep) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
