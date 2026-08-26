package mycel

import (
	"context"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
)

func (c *AdminClient) AppendActivityEvent(ctx context.Context, event *adminv1.ActivityEvent) (*adminv1.AppendActivityEventResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Activity.AppendActivityEvent(callCtx, &adminv1.AppendActivityEventRequest{Event: event})
}

func (c *AdminClient) ListActivityEvents(ctx context.Context, req *adminv1.ListActivityEventsRequest) (*adminv1.ListActivityEventsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.ListActivityEventsRequest{}
	}
	return c.Activity.ListActivityEvents(callCtx, req)
}

func (c *AdminClient) GetActivityEvent(ctx context.Context, eventID string) (*adminv1.ActivityEvent, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Activity.GetActivityEvent(callCtx, &adminv1.GetActivityEventRequest{EventId: eventID})
	if err != nil {
		return nil, err
	}
	return res.GetEvent(), nil
}
