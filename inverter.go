package foxesscloud

import (
	"context"
	"net/url"
)

type InverterService struct {
	client *Client
}

type GetInverterListOptions struct {
	Pagination `json:",inline"`
}

type Inverter struct {
	DeviceType  string `json:"deviceType"`
	HasBattery  bool   `json:"hasBattery"`
	HasPV       bool   `json:"hasPV"`
	StationName string `json:"stationName"`
	ModuleSN    string `json:"moduleSN"`
	DeviceSN    string `json:"deviceSN"`
	ProductType string `json:"productType"`
	StationID   string `json:"stationID"`
	Status      int    `json:"status"`
}

func (s *InverterService) List(ctx context.Context, opts GetInverterListOptions) (*ListResponse[Inverter], error) {
	req, err := s.client.newPostRequest(ctx, "/op/v0/device/list", opts)
	if err != nil {
		return nil, err
	}
	resp := new(wrappedListResponse[Inverter])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}

type GetInverterOptions struct {
	InverterSN string
}

type InverterDetail struct {
	DeviceType    string `json:"deviceType"`
	MasterVersion string `json:"masterVersion"`
	AFCIVersion   string `json:"afciVersion"`
	HasPV         bool   `json:"hasPV"`
	DeviceSN      string `json:"deviceSN"`
	SlaveVersion  string `json:"slaveVersion"`
	HasBattery    bool   `json:"hasBattery"`
	Function      struct {
		Scheduler bool `json:"scheduler"`
	} `json:"function"`
	HardwareVersion string `json:"hardwareVersion"`
	ManagerVersion  string `json:"managerVersion"`
	StationName     string `json:"stationName"`
	ModuleSN        string `json:"moduleSN"`
	ProductType     string `json:"productType"`
	StationID       string `json:"stationID"`
	Status          int    `json:"status"`
}

func (s *InverterService) Get(ctx context.Context, opts GetInverterOptions) (*InverterDetail, error) {
	req, err := s.client.newGetRequest(ctx, "/op/v0/device/detail", url.Values{"sn": {opts.InverterSN}})
	if err != nil {
		return nil, err
	}
	resp := new(wrappedDetailResponse[InverterDetail])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}

type GetInverterRealtimeDataOptions struct {
	InverterSN string     `json:"sn"`
	Variables  []Variable `json:"variables"`
}

type InverterRealTimeData struct {
	Datas []struct {
		Unit     string    `json:"unit"`
		Name     string    `json:"name"`
		Variable Variable  `json:"variable"`
		Value    DataFloat `json:"value"`
	} `json:"datas"`
	Time     DataTimestamp `json:"time"`
	DeviceSN string        `json:"deviceSN"`
}

func (s *InverterService) GetRealtimeData(ctx context.Context, opts GetInverterRealtimeDataOptions) (*DataListResponse[InverterRealTimeData], error) {
	req, err := s.client.newPostRequest(ctx, "/op/v0/device/real/query", opts)
	if err != nil {
		return nil, err
	}
	resp := new(wrappedDataListResponse[InverterRealTimeData])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}

type GetInverterHistoryDataOptions struct {
	InverterSN string          `json:"sn"`
	Variables  []Variable      `json:"variables"`
	Begin      *QueryTimestamp `json:"begin,omitempty"`
	End        *QueryTimestamp `json:"end,omitempty"`
}

type InverterHistoryTimeData struct {
	Datas []struct {
		Unit     string   `json:"unit"`
		Name     string   `json:"name"`
		Variable Variable `json:"variable"`
		Data     []struct {
			Time  DataTimestamp `json:"time"`
			Value DataFloat     `json:"value"`
		} `json:"data"`
	} `json:"datas"`
	DeviceSN string `json:"deviceSN"`
}

func (s *InverterService) GetHistoryData(ctx context.Context, opts GetInverterHistoryDataOptions) (*DataListResponse[InverterHistoryTimeData], error) {
	req, err := s.client.newPostRequest(ctx, "/op/v0/device/history/query", opts)
	if err != nil {
		return nil, err
	}
	resp := new(wrappedDataListResponse[InverterHistoryTimeData])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}

type GetInverterProductionReportOptions struct {
	InverterSN string     `json:"sn"`
	Year       int        `json:"year"`
	Month      *int       `json:"month,omitempty"`
	Day        *int       `json:"day,omitempty"`
	Dimension  string     `json:"dimension"`
	Variables  []Variable `json:"variables"`
}

type ProductionReport struct {
	Unit     string      `json:"unit"`
	Values   []DataFloat `json:"values"`
	Variable Variable    `json:"variable"`
}

func (s *InverterService) GetProductionReport(ctx context.Context, opts GetInverterProductionReportOptions) (*DataListResponse[ProductionReport], error) {
	req, err := s.client.newPostRequest(ctx, "/op/v0/device/report/query", opts)
	if err != nil {
		return nil, err
	}
	resp := new(wrappedDataListResponse[ProductionReport])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}
