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
	fl := make(FlagList, len(strs))
	for _, str := range strs {
		flag, err := FromString(str)
		if err != nil {
			return err
		}
		fl[flag.Name] = flag
	}
	*res = fl
	return nil
}

func (fl FlagList) MarshalJSON() ([]byte, error) {
	if len(fl) == 0 {
		return []byte("[]"), nil
	}
	strs := make([]string, 0, len(fl))
	for _, f := range fl {
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
	fl := make(FlatFlagList, len(strs))
	for _, str := range strs {
		flag, err := FlatFromString(str)
		if err != nil {
			return err
		}
		fl[flag.Name] = flag
	}
	*res = fl
	return nil
}

func (fl FlatFlagList) MarshalJSON() ([]byte, error) {
	if len(fl) == 0 {
		return []byte("[]"), nil
	}
	strs := make([]string, 0, len(fl))
	for _, f := range fl {
		strs = append(strs, f.String())
	}
	return json.Marshal(strs)
}

func (fs *FlagExpr) UnmarshalJSON(data []byte) (err error) {
	*fs, err = fromString(string(data))
	return err
}

func (fs *FlagExpr) MarshalJSON() ([]byte, error) {
	return []byte(fs.String()), nil
}
