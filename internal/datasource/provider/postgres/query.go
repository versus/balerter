package postgres

import (
	"context"
	"github.com/balerter/balerter/internal/datasource/converter"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func (m *Postgres) query(L *lua.LState) int {

	q := L.Get(1).String()

	m.logger.Debug("call postgres query", zap.String("query", q))

	ctx, ctxCancel := context.WithTimeout(context.Background(), m.timeout)
	defer ctxCancel()

	rows, err := m.db.QueryContext(ctx, q)
	if err != nil {
		m.logger.Error("error postgres query", zap.String("query", q), zap.Error(err))
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer rows.Close()

	cct, _ := rows.ColumnTypes()

	dest := make([]interface{}, 0)
	ffs := make([]func(v interface{}) lua.LValue, 0)

	for range cct {
		dest = append(dest, new([]byte))
		ffs = append(ffs, converter.FromDateBytes)
	}

	result := &lua.LTable{}

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			m.logger.Error("error scan", zap.Error(err))
			L.Push(lua.LNil)
			L.Push(lua.LString("error scan: " + err.Error()))
			return 2
		}

		row := &lua.LTable{}

		for idx, c := range cct {
			v := ffs[idx](dest[idx])
			row.RawSet(lua.LString(c.Name()), v)
		}

		result.Append(row)
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
