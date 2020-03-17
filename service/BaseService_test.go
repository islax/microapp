package service

import "testing"

func TestBaseServiceImpl_GetOrderByString(t *testing.T) {
	type args struct {
		orderByAttrs          []string
		validOrderByAttrs     []string
		orderByAttrAndDBCloum map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"TestWithNoDBCloumnMappingAndDefaultOrderByDirection", args{[]string{"col1"}, []string{"col1", "col2"}, map[string][]string{}}, "col1", false},
		{"TestWith1DBCloumnMappingAndDefaultOrderByDirection", args{[]string{"col1"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3", false},
		{"TestWith2DBCloumnMappingAndDefaultOrderByDirection", args{[]string{"col1"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3,col4", false},

		{"TestWithNoDBCloumnMappingAndOrderByDirection0", args{[]string{"col1,0"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 ASC", false},
		{"TestWithNoDBCloumnMappingAndOrderByDirectionA", args{[]string{"col1,A"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 ASC", false},
		{"TestWithNoDBCloumnMappingAndOrderByDirectionASC", args{[]string{"col1,ASC"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 ASC", false},
		{"TestWithNoDBCloumnMappingAndOrderByDirection1", args{[]string{"col1,1"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 DESC", false},
		{"TestWithNoDBCloumnMappingAndOrderByDirection_d", args{[]string{"col1,d"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 DESC", false},
		{"TestWithNoDBCloumnMappingAndOrderByDirection_desc", args{[]string{"col1,desc"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 DESC", false},

		{"TestWith1DBCloumnMappingAndOrderByDirection0", args{[]string{"col1,0"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 ASC", false},
		{"TestWith1DBCloumnMappingAndOrderByDirection_a", args{[]string{"col1,a"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 ASC", false},
		{"TestWith1DBCloumnMappingAndOrderByDirection_asc", args{[]string{"col1,asc"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 ASC", false},
		{"TestWith1DBCloumnMappingAndOrderByDirection1", args{[]string{"col1,1"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 DESC", false},
		{"TestWith1DBCloumnMappingAndOrderByDirectionD", args{[]string{"col1,d"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 DESC", false},
		{"TestWith1DBCloumnMappingAndOrderByDirectionDesc", args{[]string{"col1,desc"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3"}}}, "col3 DESC", false},

		{"TestWith2DBCloumnMappingAndOrderByDirection1", args{[]string{"col1,1"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 DESC,col4 DESC", false},
		{"TestWith2DBCloumnMappingAndOrderByDirectionD", args{[]string{"col1,D"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 DESC,col4 DESC", false},
		{"TestWith2DBCloumnMappingAndOrderByDirection_dEsC", args{[]string{"col1,dEsC"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 DESC,col4 DESC", false},
		{"TestWith2DBCloumnMappingAndOrderByDirection0", args{[]string{"col1,0"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 ASC,col4 ASC", false},
		{"TestWith2DBCloumnMappingAndOrderByDirection_a", args{[]string{"col1,a"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 ASC,col4 ASC", false},
		{"TestWith2DBCloumnMappingAndOrderByDirectionASc", args{[]string{"col1,ASc"}, []string{"col1", "col2"}, map[string][]string{"col1": []string{"col3", "col4"}}}, "col3 ASC,col4 ASC", false},

		{"TestWith2OrderByAttrNoDBCloumnMappingAndDefaultOrderByDirection", args{[]string{"col1", "col2"}, []string{"col1", "col2"}, map[string][]string{}}, "col1,col2", false},
		{"TestWith2OrderByAttrNoDBCloumnMappingAndOrderByDirections", args{[]string{"col1,1", "col2,d"}, []string{"col1", "col2"}, map[string][]string{}}, "col1 DESC,col2 DESC", false},
		{"TestWith2OrderByAttr1DBCloumnMappingAndOrderByDirections", args{[]string{"col1,a", "col2,desc"}, []string{"col1", "col2"}, map[string][]string{"col2": []string{"col3", "col4"}}}, "col1 ASC,col3 DESC,col4 DESC", false},

		{"TestWithOrderByAttrHavingTooManyCommas", args{[]string{"col1,dsc,a"}, []string{"col1", "col2"}, map[string][]string{}}, "", true},
		{"TestWithInvalidOrderByAttr", args{[]string{"col3,D"}, []string{"col1", "col2"}, map[string][]string{}}, "", true},
		{"TestWithInvalidOrderByDirection", args{[]string{"col1,dsc"}, []string{"col1", "col2"}, map[string][]string{}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &BaseServiceImpl{}
			got, err := service.CreateOrderByString(tt.args.orderByAttrs, tt.args.validOrderByAttrs, tt.args.orderByAttrAndDBCloum)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseServiceImpl.CreateOrderByString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BaseServiceImpl.CreateOrderByString() = %v, want %v", got, tt.want)
			}
		})
	}
}
