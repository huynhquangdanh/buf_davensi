package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type QueryType int

const (
	Insert QueryType = iota
	Update
	Select
	Delete
	Upsert
)

type QueryBuilder struct {
	Filters      *FilterBracket
	SelectFields []string
	TableName    string
	queryType    QueryType
	updateFields *UpdateBracket
	insertFields *InsertBracket
	conflictKeys []string
	returnFields []string
	orderBy      []string
	limit        int32
	joinClauses  []string
	groupBy      []string
}

type FilterBracket struct {
	operation string
	filters   []string
	args      []any
}

type UpdateBracket struct {
	fields  []string
	args    []any
	from    string
	fromArg any
}

type InsertBracket struct {
	fields []string
	args   []any
}

// For Filter bracket
func CreateFilterBracket(operation string) *FilterBracket {
	return &FilterBracket{
		operation: operation,
	}
}

func (filterBr *FilterBracket) SetFilter(conditions string, args ...any) *FilterBracket {
	filterBr.filters = append(filterBr.filters, conditions)
	filterBr.args = append(filterBr.args, args...)
	return filterBr
}

func (filterBr *FilterBracket) GenerateSQL() (sqlStr string, args []any) {
	if len(filterBr.filters) == 0 {
		return "", args
	}

	sqlStr = fmt.Sprintf("(%s)", strings.Join(filterBr.filters, fmt.Sprintf(" %s ", filterBr.operation)))
	return sqlStr, filterBr.args
}

// For Update bracket
func (updateBr *UpdateBracket) SetUpdate(field string, value any) *UpdateBracket {
	if updateBr.from != "" && value == nil {
		updateBr.fields = append(updateBr.fields, field)
	} else if value != nil {
		updateBr.fields = append(updateBr.fields, fmt.Sprintf("%s = ?", field))
		updateBr.args = append(updateBr.args, value)
	} else {
		updateBr.fields = append(updateBr.fields, fmt.Sprintf("%s = excluded.%s", field, field))
	}

	return updateBr
}

func (updateBr *UpdateBracket) SetUpdateFrom(from string, fromArg any) *UpdateBracket {
	updateBr.from = strings.Trim(from, " ")
	updateBr.fromArg = fromArg
	return updateBr
}

func (updateBr *UpdateBracket) IsUpdatable() bool {
	return len(updateBr.fields) > 0
}

func (updateBr *UpdateBracket) GenerateSQL() (sqlStr string, args []any) {
	if updateBr.from != "" && updateBr.fromArg != nil {
		updateBr.args = append(updateBr.args, updateBr.fromArg)
	}
	return fmt.Sprintf(
		"%s %s",
		strings.Join(updateBr.fields, ", "),
		genConditionSQL("FROM", updateBr.from),
	), updateBr.args
}

// For Insert bracket
func (insertBr *InsertBracket) SetInsertField(fields ...string) *InsertBracket {
	insertBr.fields = append(insertBr.fields, fields...)
	return insertBr
}

func (insertBr *InsertBracket) SetInsertValue(values []any) (*InsertBracket, error) {
	if len(insertBr.fields) != len(values) {
		return insertBr, fmt.Errorf("values '%v' must match with fields '%v' to insert", values, insertBr.fields)
	}
	insertBr.args = append(insertBr.args, values...)
	return insertBr, nil
}

func (insertBr *InsertBracket) GenerateSQL() (sqlStr string, args []any) {
	numberRows := len(insertBr.args) / len(insertBr.fields)
	argsValues := []string{}
	for i := 0; i < numberRows; i++ {
		argsValues = append(
			argsValues,
			strings.Join(
				strings.Split(
					strings.Repeat("?", len(insertBr.fields)),
					"",
				),
				", ",
			),
		)
	}
	sqlStr = fmt.Sprintf(
		"(%s) VALUES (%s)",
		strings.Join(insertBr.fields, ", "),
		strings.Join(argsValues, "), ("),
	)

	return sqlStr, insertBr.args
}

func CreateQueryBuilder(
	queryType QueryType,
	tableName string,
) *QueryBuilder {
	return &QueryBuilder{
		TableName:    tableName,
		queryType:    queryType,
		Filters:      CreateFilterBracket("AND"),
		updateFields: &UpdateBracket{},
		insertFields: &InsertBracket{},
	}
}

func (qb *QueryBuilder) Select(field string) *QueryBuilder {
	qb.SelectFields = append(qb.SelectFields, field)
	return qb
}

func (qb *QueryBuilder) SuperSelect(field string) *QueryBuilder {
	listField := strings.Split(field, ",")
	newField := ""
	for _, strField := range listField {
		strField = qb.TableName + "." + strings.TrimSpace(strField)
		newField += strField + ","
	}
	newField = strings.TrimSuffix(newField, ",")
	qb.SelectFields = append(qb.SelectFields, newField)
	return qb
}

func (qb *QueryBuilder) Where(conditions string, args ...any) *QueryBuilder {
	if len(conditions) > 0 {
		qb.Filters.SetFilter(conditions, args...)
	}
	return qb
}

func (qb *QueryBuilder) OrderBy(order string) *QueryBuilder {
	if order != "" {
		qb.orderBy = append(qb.orderBy, order)
	} else {
		qb.orderBy = []string{}
	}
	return qb
}

func (qb *QueryBuilder) Limit(limit int32) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) SetUpdate(field string, value any) *QueryBuilder {
	qb.updateFields.SetUpdate(field, value)
	return qb
}

func (qb *QueryBuilder) IsUpdatable() bool {
	return qb.updateFields.IsUpdatable()
}

func (qb *QueryBuilder) SetUpdateFrom(from string, fromArg any) *QueryBuilder {
	qb.updateFields.SetUpdateFrom(from, fromArg)
	return qb
}

func (qb *QueryBuilder) SetInsertField(field ...string) *QueryBuilder {
	qb.insertFields.SetInsertField(field...)
	return qb
}

func (qb *QueryBuilder) SetReturnFields(fields ...string) *QueryBuilder {
	qb.returnFields = append(qb.returnFields, fields...)
	return qb
}

func (qb *QueryBuilder) SetInsertValues(values []any) (*QueryBuilder, error) {
	_, err := qb.insertFields.SetInsertValue(values)
	return qb, err
}

func (qb *QueryBuilder) Join(joinClause string) *QueryBuilder {
	qb.joinClauses = append(qb.joinClauses, joinClause)
	return qb
}

func (qb *QueryBuilder) SuperJoin(joinClause, addFieldSelect, tableJoin string) *QueryBuilder {
	listField := strings.Split(addFieldSelect, ",")
	newField := ""
	for _, strField := range listField {
		strField = tableJoin + "." + strings.TrimSpace(strField)
		newField += strField + ","
	}
	newField = strings.TrimSuffix(newField, ",")
	qb.SelectFields = append(qb.SelectFields, newField)
	qb.joinClauses = append(qb.joinClauses, joinClause)
	return qb
}

func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	qb.groupBy = fields
	return qb
}

func (qb *QueryBuilder) OnConflict(fields ...string) *QueryBuilder {
	qb.conflictKeys = append(qb.conflictKeys, fields...)
	return qb
}

func (qb *QueryBuilder) GenerateSQL() (sqlStr string, args []any, sel string) {
	if (qb.queryType == Insert || qb.queryType == Update) && len(qb.returnFields) == 0 {
		qb.SetReturnFields("*")
	}

	switch qb.queryType {
	case Select:
		filterSQL, filterArgs := qb.Filters.GenerateSQL()

		sel = fmt.Sprintf("for %s filter %v", filterSQL, filterArgs)
		args = filterArgs
		sqlStr = fmt.Sprintf(
			"SELECT %s FROM %s %s %s %s %s %s",
			qb.getSelectFields(),
			qb.TableName,
			strings.Join(qb.joinClauses, " "),
			strings.Join(qb.groupBy, " "),
			genConditionSQL("WHERE", filterSQL),
			genConditionSQL("ORDER BY", strings.Join(qb.orderBy, ", ")),
			genConditionSQL("LIMIT", qb.getLimit()),
		)
	case Insert:
		insertSQL, insertArgs := qb.insertFields.GenerateSQL()
		args = insertArgs

		if len(qb.conflictKeys) > 0 && qb.updateFields.IsUpdatable() {
			updateSQL, updateArgs := qb.updateFields.GenerateSQL()
			insertSQL = fmt.Sprintf(
				"%s ON CONFLICT (%s) DO UPDATE SET %s",
				insertSQL,
				strings.Join(qb.conflictKeys, ", "),
				updateSQL,
			)
			args = append(args, updateArgs...)
		}
		sel = fmt.Sprintf("for %s insert %v", insertSQL, insertArgs)
		sqlStr = fmt.Sprintf(
			"INSERT INTO %s%s %s",
			qb.TableName,
			insertSQL,
			genConditionSQL("RETURNING", strings.Join(qb.returnFields, ", ")),
		)
	case Upsert:
		insertSQL, insertArgs := qb.insertFields.GenerateSQL()

		sel = fmt.Sprintf("for %s insert %v", insertSQL, insertArgs)
		args = insertArgs
		sqlStr = fmt.Sprintf(
			"UPSERT INTO %s%s %s",
			qb.TableName,
			insertSQL,
			genConditionSQL("RETURNING", strings.Join(qb.returnFields, ", ")),
		)
	case Update:
		filterSQL, filterArgs := qb.Filters.GenerateSQL()
		updateSQL, updateArgs := qb.updateFields.GenerateSQL()

		sel = fmt.Sprintf("update %s value %v by filter %s value %v", updateSQL, updateArgs, filterSQL, filterArgs)
		args = append(args, updateArgs...)
		args = append(args, filterArgs...)
		sqlStr = fmt.Sprintf(
			"UPDATE %s SET %s %s %s;",
			qb.TableName,
			updateSQL,
			genConditionSQL("WHERE", filterSQL),
			genConditionSQL("RETURNING", strings.Join(qb.returnFields, ", ")),
		)
	case Delete:
		filterSQL, filterArgs := qb.Filters.GenerateSQL()

		sel = fmt.Sprintf("for %s filter %v", filterSQL, filterArgs)
		args = filterArgs
		sqlStr = fmt.Sprintf(
			"DELETE FROM %s %s %s",
			qb.TableName,
			genConditionSQL("WHERE", filterSQL),
			genConditionSQL("RETURNING", strings.Join(qb.returnFields, ", ")),
		)
	}

	return replaceSQLArgs(sqlStr), args, sel
}

func (qb *QueryBuilder) getLimit() string {
	if qb.limit > 0 {
		return fmt.Sprintf("%d", qb.limit)
	}
	return ""
}

func (qb *QueryBuilder) getSelectFields() string {
	if len(qb.SelectFields) > 0 {
		return strings.Join(qb.SelectFields, ", ")
	}
	return "*"
}

func genConditionSQL(condition, values string) (where string) {
	if len(values) > 0 {
		where = fmt.Sprintf("%s %s", condition, values)
	}
	return where
}

func replaceSQLArgs(sql string) string {
	argIndex := 1
	regex := regexp.MustCompile(`(\?|(\$\d+)|\$)+`)

	sql = regex.ReplaceAllStringFunc(sql, func(match string) string {
		if regex.MatchString(match) {
			indexString := fmt.Sprintf("$%s", strconv.Itoa(argIndex))
			argIndex++
			return indexString
		}
		return match
	})

	return sql
}
