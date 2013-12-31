package query

import (
	"pilosa/db"
	"github.com/nu7hatch/gouuid"
)

type QueryInput interface{}

type QueryResults struct {
	Data interface{}
}

type Query struct {
	Operation string
	Inputs    []QueryInput //"strconv"
	// Represents a parsed query. Inputs can be Query or Bitmap objects
	// Maybe Bitmap and Query objects should have different fields to avoid using interface{}
	Profile_id int
}

func QueryPlanForPQL(database *db.Database, pql string) *QueryPlan {
	//spew.Dump("EXECUTE")
	//spew.Dump(pql)
	tokens := Lex(pql)
	//spew.Dump(tokens)

	query_parser := QueryParser{}
	query, err := query_parser.Parse(tokens)
	if err != nil {
		panic(err)
	}
	//spew.Dump(query)

	// switch on different query types:
	//if query.Operation == "set" {
	//spew.Dump("SET!!")
	//}

	query_planner := QueryPlanner{Database: database}
	id, _ := uuid.NewV4()
	destination := db.Process{}
	query_plan := query_planner.Plan(query, id, &destination)
	return query_plan
}
