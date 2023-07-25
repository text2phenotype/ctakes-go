package fsm

import (
	"text2phenotype.com/fdl/utils"
	"errors"
	"path"
	"reflect"
)

type Params struct {
	DosageFSM        DosageFSMParams              `drug:"dosage_fsm"`
	ChangeStatusFSM  DrugChangeStatusParams       `drug:"drug_change_status_fsm"`
	DurationFSM      DurationFSMParams            `drug:"duration_fsm"`
	FormFSM          FormFSMParams                `drug:"form_fsm"`
	FractionFSM      FractionStrengthFSMParams    `drug:"fraction_strength_fsm"`
	FrequencyFSM     FrequencyFSMParams           `drug:"frequency_fsm"`
	FrequencyUnitFSM FrequencyUnitFSMParams       `drug:"frequency_unit_fsm"`
	RangeFSM         RangeStrengthFSMParams       `drug:"range_strength_fsm"`
	RouteFSM         RouteFSMParams               `drug:"route_fsm"`
	StrengthFSM      StrengthFSMParams            `drug:"strength_fsm"`
	StrengthUnitFSM  StrengthUnitFSMParams        `drug:"strength_unit_fsm"`
	SubMedSectionFSM SubSectionIndicatorFSMParams `drug:"subsection_indicator_fsm"`
	SuffixFSM        SuffixStrengthFSMParams      `drug:"suffix_strength_fsm"`
	TimeFSM          TimeFSMParams                `drug:"time_fsm"`
}

func LoadDrugFSMExtractorParams(resPath string) (Params, error) {
	root := path.Join(resPath, "drug_ner/fsm")

	var params Params
	paramsVal := reflect.ValueOf(&params)
	paramsTp := reflect.TypeOf(&params)
	fieldsCount := paramsTp.Elem().NumField()
	for fieldIdx := 0; fieldIdx < fieldsCount; fieldIdx++ {
		f := paramsTp.Elem().Field(fieldIdx)
		dirName, ok := f.Tag.Lookup("drug")
		if !ok {
			continue
		}

		paramsDirPath := path.Join(root, dirName)

		paramFieldVal := paramsVal.Elem().Field(fieldIdx)
		paramsCount := f.Type.NumField()
		for paramFieldIdx := 0; paramFieldIdx < paramsCount; paramFieldIdx++ {
			pf := f.Type.Field(paramFieldIdx)
			if pf.Type.Key().Kind() != reflect.String {
				return params, errors.New("drug params: field key type is not supported")
			}

			if pf.Type.Elem().Kind() != reflect.String && pf.Type.Elem().Kind() != reflect.Bool {
				return params, errors.New("drug params: Field value type is not supported")
			}

			paramFilePath, ok := pf.Tag.Lookup("drug")

			val := paramFieldVal.Field(paramFieldIdx)

			if !ok {
				val.Set(reflect.MakeMap(pf.Type))
				continue
			}

			fullPath := path.Join(paramsDirPath, paramFilePath)
			if pf.Type.Elem().Kind() == reflect.String {
				m, err := utils.ReadMap(fullPath)
				if err != nil {
					continue
				}
				val.Set(reflect.ValueOf(m))

			} else {
				s, err := utils.ReadSet(fullPath)
				if err != nil {
					continue
				}

				val.Set(reflect.ValueOf(s))
			}

		}

	}

	return params, nil
}
