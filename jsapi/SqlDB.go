package jsapi

import (
	js "github.com/dop251/goja"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlDB struct {
	VM      *js.Runtime
	Proto   *js.Object
	DataDir string
}

func (s *SqlDB) Init() {
	cVal := s.VM.ToValue(s.constructor)
	s.Proto = cVal.(*js.Object).Get("prototype").(*js.Object)

	s.VM.Set("DB", cVal)
}

func (s *SqlDB) constructor(call js.ConstructorCall) *js.Object {
	db, err := gorm.Open(sqlite.Open("file:"+s.DataDir+"/ext.db"), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	inst := s.VM.CreateObject(s.Proto)

	inst.Set("exec", func(call js.FunctionCall) js.Value {
		query := call.Argument(0).String()
		args := []any{}

		if len(call.Arguments) > 1 {
			for _, arg := range call.Arguments[1:] {
				args = append(args, arg.Export())
			}
		}

		tx := db.Exec(query, args...)

		if tx.Error != nil {
			s.VM.Interrupt(tx.Error)
			return js.Null()
		}

		obj := make(map[string]any)

		obj["rowsAffected"] = tx.RowsAffected

		return s.VM.ToValue(obj)
	})

	inst.Set("query", func(call js.FunctionCall) js.Value {
		query := call.Argument(0).String()
		args := []any{}

		if len(call.Arguments) > 1 {
			for _, arg := range call.Arguments[1:] {
				args = append(args, arg.Export())
			}
		}

		results := make([]map[string]any, 0)

		tx := db.Raw(query, args...).Scan(&results)

		if tx.Error != nil {
			s.VM.Interrupt(tx.Error)
			return js.Null()
		}

		return s.VM.ToValue(results)
	})

	return inst
}
