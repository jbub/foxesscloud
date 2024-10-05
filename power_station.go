package foxesscloud

import (
	"context"
	"net/url"
)

type PowerStationService struct {
	client *Client
}

type GetPowerStationOptions struct {
	StationID string
}

type PowerStationDetail struct {
	Country   string `json:"country"`
	Address   string `json:"address"`
	Installer struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"installer"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
	Postcode    string `json:"postcode"`
	StationName string `json:"stationName"`
	User        struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
	Modules []struct {
		ModuleSN string `json:"moduleSN"`
		DeviceSN string `json:"deviceSN"`
	} `json:"modules"`
	Capacity   float64       `json:"capacity"`
	CreateDate DataTimestamp `json:"createDate"`
}

func (s *PowerStationService) Get(ctx context.Context, opts GetPowerStationOptions) (*PowerStationDetail, error) {
	req, err := s.client.newGetRequest(ctx, "/op/v0/plant/detail", url.Values{"id": {opts.StationID}})
	if err != nil {
		return nil, err
	}
	resp := new(wrappedDetailResponse[PowerStationDetail])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}

type GetPowerStationListOptions struct {
	Pagination `json:",inline"`
}

type PowerStation struct {
	Name         string `json:"name"`
	IanaTimezone string `json:"ianaTimezone"`
	StationID    string `json:"stationID"`
}

func (s *PowerStationService) List(ctx context.Context, opts GetPowerStationListOptions) (*ListResponse[PowerStation], error) {
	req, err := s.client.newPostRequest(ctx, "/op/v0/plant/list", opts)
	if err != nil {
		return nil, err
	}
	resp := new(wrappedListResponse[PowerStation])
	if _, err = s.client.do(req, resp); err != nil {
		return nil, err
	}
	return resp.unwrap(), nil
}
