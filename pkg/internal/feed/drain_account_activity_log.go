package feed

// DrainAccountActivityHistoryLog cycle through GetAccountsActivityLogResponse and adapts the filter while there is a next page
func DrainAccountActivityHistoryLog(ctx context.Context, req *GetAccountsActivityLogRequestParams, feedFn func(*GetAccountsActivityLogResponse) error) error {
	for {
		res, err := a.ListOrganisationActivityLog(ctx, req)
		if err != nil {
			return err
		}

		err = feedFn(res)
		if err != nil {
			return err
		}

		if res.NextPageToken != "" {
			req.PageToken = res.NextPageToken
		} else {
			break
		}
	}
	return nil
}
