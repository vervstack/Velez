package container_manager_v1

//
//const (
//	matreshkaConfigLabel = "MATRESHKA_CONFIG_ENABLED"
//)
//
//func (c *ContainerManager) LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error) {
//	err = c.normalizeCreateRequest(req)
//	if err != nil {
//		return "", errors.Wrap(err, "error normalizing create request")
//	}
//
//	image, err := dockerutils.PullImage(ctx, c.docker, req.ImageName, false)
//	if err != nil {
//		return "", errors.Wrap(err, "error pulling image")
//	}
//
//	var cont *types.ContainerJSON
//
//	if image.Labels[matreshkaConfigLabel] == "true" {
//		// TODO Do create verv here
//	}
//
//	cont, err = c.deployManager.Create(ctx, req)
//	if err != nil {
//		return "", errors.Wrap(err, "error creating container")
//	}
//
//	err = c.deployManager.Start(ctx, cont.ID)
//	if err != nil {
//		return "", errors.Wrap(err, "error starting container")
//	}
//
//	err = c.deployManager.Healthcheck(ctx, cont.ID, req.Healthcheck)
//
//	if req.Healthcheck != nil {
//		err = c.doHealthcheck(ctx, cont.ID, req.Healthcheck)
//		if err != nil {
//			return "", errors.Wrap(err, "error during healthcheck")
//		}
//	}
//
//	return cont.ID, nil
//}
//
//func (c *ContainerManager) normalizeCreateRequest(req *velez_api.CreateSmerd_Request) error {
//	if req.Settings == nil {
//		req.Settings = &velez_api.Container_Settings{}
//	}
//
//	if req.Hardware == nil {
//		req.Hardware = &velez_api.Container_Hardware{}
//	}
//
//	if req.Env == nil {
//		req.Env = make(map[string]string)
//	}
//
//	for _, p := range req.Settings.Ports {
//		if p.Host == 0 {
//			var err error
//			p.Host, err = c.portManager.GetPort()
//			if err != nil {
//				return errors.Wrap(err, "error getting host port")
//			}
//		} else {
//			err := c.portManager.LockPorts(req.Settings.Ports)
//			if err != nil {
//				return errors.Wrap(err, "error locking ports for container")
//			}
//		}
//
//	}
//
//	if req.Labels == nil {
//		req.Labels = make(map[string]string)
//	}
//
//	req.Labels[CreatedWithVelezLabel] = "true"
//
//	return nil
//}
//
//// Move to deploy manager
//func (c *ContainerManager) doHealthcheck(
//	ctx context.Context,
//	containerId string,
//	hc *velez_api.Container_Healthcheck,
//) error {
//	errC := make(chan error)
//
//	go func() {
//		defer close(errC)
//
//		for i := uint32(0); i < hc.Retries; i++ {
//			time.Sleep(time.Duration(hc.IntervalSecond) * time.Second)
//
//			cont, err := c.docker.InspectContainer(ctx, containerId)
//			if err != nil {
//				errC <- err
//				return
//			}
//			if cont.State.Health == nil {
//				continue
//			}
//
//			if cont.State.Status == "running" {
//				errC <- nil
//				return
//			}
//		}
//	}()
//
//	err := <-errC
//	if err != nil {
//		return errors.Wrap(err, "error during healthcheck")
//	}
//	return nil
//
//}
