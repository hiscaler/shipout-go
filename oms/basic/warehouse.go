package basic

import (
	"errors"
	"github.com/hiscaler/shipout-go"
	jsoniter "github.com/json-iterator/go"
)

// 仓库
// https://open.shipout.com/portal/zh/api/base-info.html

type Warehouse struct {
	OrgId             string `json:"orgId"`             // 仓库机构ID
	TimeZone          string `json:"timeZone"`          // 仓库时区
	WarehouseAddr1    string `json:"warehouseAddr1"`    // 仓库地址1
	WarehouseAddr2    string `json:"warehouseAddr2"`    // 仓库地址2
	WarehouseCity     string `json:"warehouseCity"`     // 仓库所在城市
	WarehouseContacts string `json:"warehouseContacts"` // 仓库联系人
	WarehouseCountry  string `json:"warehouseCountry"`  // 仓库所在国家
	WarehouseEmail    string `json:"warehouseEmail"`    // 仓库联系Email
	WarehouseId       string `json:"warehouseId"`       // 仓库编号
	WarehouseName     string `json:"warehouseName"`     // 仓库名称
	WarehousePhone    string `json:"warehousePhone"`    // 仓库联系电话
	WarehouseProvince string `json:"warehouseProvince"` // 仓库所在州
	WarehouseZipCode  string `json:"warehouseZipCode"`  // 仓库邮编
}

type WarehousesQueryParams struct {
	Name string `json:"name"`
}

func (m WarehousesQueryParams) Validate() error {
	return nil
}

func (s service) Warehouses(params WarehousesQueryParams) (items []Warehouse, isLastPage bool, err error) {
	if err = params.Validate(); err != nil {
		return
	}

	res := struct {
		shipout.NormalResponse
		Data []Warehouse `json:"data"`
	}{}
	qp := make(map[string]string, 0)
	if params.Name != "" {
		qp["name"] = params.Name
	}
	resp, err := s.shipOut.Client.R().
		SetQueryParams(qp).
		Get("/open-api/oms/info/warehouse/list")
	if err != nil {
		return
	}

	if resp.IsSuccess() {
		if err = shipout.ErrorWrap(res.Result, res.Message); err == nil {
			if err = jsoniter.Unmarshal(resp.Body(), &res); err == nil {
				items = res.Data
			}
		}
	} else {
		if e := jsoniter.Unmarshal(resp.Body(), &res); e == nil {
			err = shipout.ErrorWrap(res.Result, res.Message)
		} else {
			err = errors.New(resp.Status())
		}
	}
	return
}
