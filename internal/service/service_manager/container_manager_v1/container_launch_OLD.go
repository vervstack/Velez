package container_manager_v1

//func createVerv(ctx context.Context, req *velez_api.CreateSmerd_Request) (*types.ContainerJSON, error) {
//	matreshkaConfig, err := c.configManager.GetFromApi(ctx, req.GetName())
//	if err != nil {
//		return nil, errors.Wrap(err, "error getting matreshka config from matreshka api")
//	}
//
//	for _, srv := range matreshkaConfig.Servers {
//		req.Settings.Ports = append(req.Settings.Ports,
//			&velez_api.PortBindings{
//				Container: uint32(srv.GetPort()),
//				Protoc:    velez_api.PortBindings_tcp,
//			})
//	}
//
//	// Create network for smerd
//	{
//		err = dockerutils.CreateNetworkSoft(ctx, c.docker, req.GetName())
//		if err != nil {
//			return nil, errors.Wrap(err, "error creating network for service")
//		}
//
//		req.Settings.Networks = append(req.Settings.Networks,
//			&velez_api.NetworkBind{
//				NetworkName: req.GetName(),
//				Aliases:     []string{req.GetName()},
//			},
//		)
//	}
//
//	// Verv-Env variables
//	{
//		req.Env[matreshka.VervName] = req.GetName()
//		//req.Env[matreshka.ApiURL] = matreshkaUrl
//	}
//
//	cont, err := c.createSimple(ctx, req)
//	if err != nil {
//		return nil, errors.Wrap(err, "error creating pre container")
//	}
//
//	return cont, nil
//}
