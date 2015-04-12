package flag

import "encoding/json"

func (fl FlagList) UnmarshalJSON(data []byte) error {
	var strs []string
	err := json.Unmarshal(data, &strs)
	if err != nil {
		return err
	}
	for _, str := range strs {
		flag, err := FromString(str)
		if err != nil {
			return err
		}
		fl[flag.Name] = flag
	}
	return nil
}

func (fl FlagList) MarshalJSON() ([]byte, error) {
	return []byte(fl.String()), nil
}

func (fl FlatFlagList) UnmarshalJSON(data []byte) error {
	var strs []string
	err := json.Unmarshal(data, &strs)
	if err != nil {
		return err
	}
	for _, str := range strs {
		flag, err := FlatFromString(str)
		if err != nil {
			return err
		}
		fl[flag.Name] = flag
	}
	return nil
}

func (fl FlatFlagList) MarshalJSON() ([]byte, error) {
	return []byte(fl.String()), nil
}
