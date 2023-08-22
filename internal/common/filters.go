package common

import (
	"fmt"

	pbCommon "davensi.com/core/gen/common"

	"davensi.com/core/internal/util"
)

func setRangeCondition(
	expression, field string,
	value *pbCommon.UInt32Boundary,
	filterBracket *util.FilterBracket,
) {
	switch value.GetBoundary().(type) {
	case *pbCommon.UInt32Boundary_Incl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s= ?", field, expression),
			value.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl,
		)
	case *pbCommon.UInt32Boundary_Excl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s ?", field, expression),
			value.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl,
		)
	}
}

func GetDecimalsFB(
	list *pbCommon.UInt32ValueList,
	field string,
) *util.FilterBracket {
	filterBracket := util.CreateFilterBracket("OR")

	for _, v := range list.GetList() {
		switch v.GetSelect().(type) {
		case *pbCommon.UInt32Values_Single:
			filterBracket.SetFilter(
				fmt.Sprintf("%s = ?", field),
				v.GetSelect().(*pbCommon.UInt32Values_Single).Single,
			)
		case *pbCommon.UInt32Values_Range:
			from := v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From
			to := v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To
			if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil && v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil {
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
			}
		}
	}

	return filterBracket
}
