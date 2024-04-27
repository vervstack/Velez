package grpc

//func (a *Api) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
//	err := req.ValidateAll()
//	if err != nil {
//		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "invalid request. Must match \"lowercase/lowercase:v0.0.1\"").Error())
//	}
//
//	id, err := a.containerManager.LaunchSmerd(ctx, req)
//	if err != nil {
//		return nil, errors.Wrapf(err, "error launching smerd")
//	}
//
//	smerds, err := a.containerManager.ListSmerds(ctx,
//		&velez_api.ListSmerds_Request{
//			Id: &id,
//		})
//	if err != nil {
//		return nil, err
//	}
//
//	if len(smerds.Smerds) == 0 {
//		return nil, status.Error(codes.NotFound, "created but couldn't find it")
//	}
//
//	return smerds.Smerds[0], nil
//}
