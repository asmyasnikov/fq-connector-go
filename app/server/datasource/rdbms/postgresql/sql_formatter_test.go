package postgresql

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	ydb "github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	api_service_protos "github.com/ydb-platform/fq-connector-go/api/service/protos"
	rdbms_utils "github.com/ydb-platform/fq-connector-go/app/server/datasource/rdbms/utils"
	"github.com/ydb-platform/fq-connector-go/common"
)

func TestMakeReadSplitQuery(t *testing.T) {
	type testCase struct {
		testName    string
		selectReq   *api_service_protos.TSelect
		outputQuery string
		outputArgs  []any
		err         error
	}

	logger := common.NewTestLogger(t)
	formatter := NewSQLFormatter()

	tcs := []testCase{
		{
			testName: "empty_table_name",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "",
				},
				What: &api_service_protos.TSelect_TWhat{},
			},
			outputQuery: "",
			outputArgs:  nil,
			err:         common.ErrEmptyTableName,
		},
		{
			testName: "empty_no columns",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: &api_service_protos.TSelect_TWhat{},
			},
			outputQuery: `SELECT 0 FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "select_col",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: &api_service_protos.TSelect_TWhat{
					Items: []*api_service_protos.TSelect_TWhat_TItem{
						{
							Payload: &api_service_protos.TSelect_TWhat_TItem_Column{
								Column: &ydb.Column{
									Name: "col",
									Type: rdbms_utils.NewPrimitiveType(ydb.Type_INT32),
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col" FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "is_null",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_IsNull{
							IsNull: &api_service_protos.TPredicate_TIsNull{
								Value: rdbms_utils.NewColumnExpression("col1"),
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE ("col1" IS NULL)`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "is_not_null",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_IsNotNull{
							IsNotNull: &api_service_protos.TPredicate_TIsNotNull{
								Value: rdbms_utils.NewColumnExpression("col2"),
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE ("col2" IS NOT NULL)`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "bool_column",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_BoolExpression{
							BoolExpression: &api_service_protos.TPredicate_TBoolExpression{
								Value: rdbms_utils.NewColumnExpression("col2"),
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE "col2"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "complex_filter",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_Disjunction{
							Disjunction: &api_service_protos.TPredicate_TDisjunction{
								Operands: []*api_service_protos.TPredicate{
									{
										Payload: &api_service_protos.TPredicate_Negation{
											Negation: &api_service_protos.TPredicate_TNegation{
												Operand: &api_service_protos.TPredicate{
													Payload: &api_service_protos.TPredicate_Comparison{
														Comparison: &api_service_protos.TPredicate_TComparison{
															Operation:  api_service_protos.TPredicate_TComparison_LE,
															LeftValue:  rdbms_utils.NewColumnExpression("col2"),
															RightValue: rdbms_utils.NewInt32ValueExpression(42),
														},
													},
												},
											},
										},
									},
									{
										Payload: &api_service_protos.TPredicate_Conjunction{
											Conjunction: &api_service_protos.TPredicate_TConjunction{
												Operands: []*api_service_protos.TPredicate{
													{
														Payload: &api_service_protos.TPredicate_Comparison{
															Comparison: &api_service_protos.TPredicate_TComparison{
																Operation:  api_service_protos.TPredicate_TComparison_NE,
																LeftValue:  rdbms_utils.NewColumnExpression("col1"),
																RightValue: rdbms_utils.NewUint64ValueExpression(0),
															},
														},
													},
													{
														Payload: &api_service_protos.TPredicate_IsNull{
															IsNull: &api_service_protos.TPredicate_TIsNull{
																Value: rdbms_utils.NewColumnExpression("col3"),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE ((NOT ("col2" <= $1)) OR (("col1" <> $2) AND ("col3" IS NULL)))`,
			outputArgs:  []any{int32(42), uint64(0)},
			err:         nil,
		},
		{
			testName: "unsupported_predicate",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_Between{
							Between: &api_service_protos.TPredicate_TBetween{
								Value:    rdbms_utils.NewColumnExpression("col2"),
								Least:    rdbms_utils.NewColumnExpression("col1"),
								Greatest: rdbms_utils.NewColumnExpression("col3"),
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "unsupported_type",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_Comparison{
							Comparison: &api_service_protos.TPredicate_TComparison{
								Operation:  api_service_protos.TPredicate_TComparison_EQ,
								LeftValue:  rdbms_utils.NewColumnExpression("col2"),
								RightValue: rdbms_utils.NewTextValueExpression("text"),
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "partial_filter_removes_and",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_Conjunction{
							Conjunction: &api_service_protos.TPredicate_TConjunction{
								Operands: []*api_service_protos.TPredicate{
									{
										Payload: &api_service_protos.TPredicate_Comparison{
											Comparison: &api_service_protos.TPredicate_TComparison{
												Operation:  api_service_protos.TPredicate_TComparison_EQ,
												LeftValue:  rdbms_utils.NewColumnExpression("col1"),
												RightValue: rdbms_utils.NewInt32ValueExpression(32),
											},
										},
									},
									{
										// Not supported
										Payload: &api_service_protos.TPredicate_Comparison{
											Comparison: &api_service_protos.TPredicate_TComparison{
												Operation:  api_service_protos.TPredicate_TComparison_EQ,
												LeftValue:  rdbms_utils.NewColumnExpression("col2"),
												RightValue: rdbms_utils.NewTextValueExpression("text"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE ("col1" = $1)`,
			outputArgs:  []any{int32(32)},
			err:         nil,
		},
		{
			testName: "partial_filter",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: rdbms_utils.NewDefaultWhat(),
				Where: &api_service_protos.TSelect_TWhere{
					FilterTyped: &api_service_protos.TPredicate{
						Payload: &api_service_protos.TPredicate_Conjunction{
							Conjunction: &api_service_protos.TPredicate_TConjunction{
								Operands: []*api_service_protos.TPredicate{
									{
										Payload: &api_service_protos.TPredicate_Comparison{
											Comparison: &api_service_protos.TPredicate_TComparison{
												Operation:  api_service_protos.TPredicate_TComparison_EQ,
												LeftValue:  rdbms_utils.NewColumnExpression("col1"),
												RightValue: rdbms_utils.NewInt32ValueExpression(32),
											},
										},
									},
									{
										// Not supported
										Payload: &api_service_protos.TPredicate_Comparison{
											Comparison: &api_service_protos.TPredicate_TComparison{
												Operation:  api_service_protos.TPredicate_TComparison_EQ,
												LeftValue:  rdbms_utils.NewColumnExpression("col2"),
												RightValue: rdbms_utils.NewTextValueExpression("text"),
											},
										},
									},
									{
										Payload: &api_service_protos.TPredicate_IsNull{
											IsNull: &api_service_protos.TPredicate_TIsNull{
												Value: rdbms_utils.NewColumnExpression("col3"),
											},
										},
									},
									{
										Payload: &api_service_protos.TPredicate_IsNotNull{
											IsNotNull: &api_service_protos.TPredicate_TIsNotNull{
												Value: rdbms_utils.NewColumnExpression("col4"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "col0", "col1" FROM "tab" WHERE (("col1" = $1) AND ("col3" IS NULL) AND ("col4" IS NOT NULL))`,
			outputArgs:  []any{int32(32)},
			err:         nil,
		},
		{
			testName: "negative_sql_injection_by_table",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: `information_schema.columns; DROP TABLE information_schema.columns`,
				},
				What: &api_service_protos.TSelect_TWhat{},
			},
			outputQuery: `SELECT 0 FROM "information_schema.columns; DROP TABLE information_schema.columns"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "negative_sql_injection_by_col",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: &api_service_protos.TSelect_TWhat{
					Items: []*api_service_protos.TSelect_TWhat_TItem{
						{
							Payload: &api_service_protos.TSelect_TWhat_TItem_Column{
								Column: &ydb.Column{
									Name: `0; DROP TABLE information_schema.columns`,
									Type: rdbms_utils.NewPrimitiveType(ydb.Type_INT32),
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "0; DROP TABLE information_schema.columns" FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
		{
			testName: "negative_sql_injection_fake_quotes",
			selectReq: &api_service_protos.TSelect{
				From: &api_service_protos.TSelect_TFrom{
					Table: "tab",
				},
				What: &api_service_protos.TSelect_TWhat{
					Items: []*api_service_protos.TSelect_TWhat_TItem{
						{
							Payload: &api_service_protos.TSelect_TWhat_TItem_Column{
								Column: &ydb.Column{
									Name: `0"; DROP TABLE information_schema.columns;`,
									Type: rdbms_utils.NewPrimitiveType(ydb.Type_INT32),
								},
							},
						},
					},
				},
			},
			outputQuery: `SELECT "0""; DROP TABLE information_schema.columns;" FROM "tab"`,
			outputArgs:  []any{},
			err:         nil,
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.testName, func(t *testing.T) {
			output, outputArgs, err := rdbms_utils.MakeReadSplitQuery(logger, formatter, tc.selectReq)
			require.Equal(t, tc.outputQuery, output)
			require.Equal(t, tc.outputArgs, outputArgs)

			if tc.err != nil {
				require.True(t, errors.Is(err, tc.err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
