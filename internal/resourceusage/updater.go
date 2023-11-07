package resourceusage

/*
type Updater interface {
	UpdateResourceUsage(ctx context.Context) error
}

type updater struct {
	k8sClient   k8s.Client
	teamsClient teams.Client
	querier     gensql.Querier
	log         logrus.FieldLogger
}

func NewUpdater(k8sClient k8s.Client, teamsClient teams.Client, querier gensql.Querier, log logrus.FieldLogger) Updater {
	return &updater{
		k8sClient:   k8sClient,
		teamsClient: teamsClient,
		querier:     querier,
		log:         log,
	}
}

func (u *updater) UpdateResourceUsage(ctx context.Context) error {
	u.log.Debugf("fetching all teams")
	allTeams, err := u.teamsClient.GetTeams(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch teams: %w", err)
	}

	resourceTypes := []model.ResourceType{
		model.ResourceTypeCPU,
		model.ResourceTypeMemory,
	}

	start := time.Now()
	appsUpdated := 0
	for _, env := range u.k8sClient.ClusterNames {
		log := u.log.WithField("env", env)

		for _, team := range allTeams {
			log = log.WithFields(logrus.Fields{"team": team.Slug})
			log.Debugf("fetching apps")

			teamApps, err := u.k8sClient.AppsInEnv(ctx, team.Slug, env)
			if err != nil {
				log.WithError(err).Errorf("unable to fetch apps")
				continue
			}

			// fetch resource usage for each app
			for _, app := range teamApps {
				log = log.WithField("app", app.Name)

				for _, resourceType := range resourceTypes {
					log = log.WithField("resourceType", resourceType)

					lastLowresDate, err := u.querier.LastLowResResourceUtilizationDateForApp(ctx, gensql.LastLowResResourceUtilizationDateForAppParams{
						Env:          env,
						ResourceType: gensql.ResourceType(resourceType),
						Team:         team.Slug,
						App:          app.Name,
					})
					if err != nil {
						log.WithError(err).Errorf("unable to fetch latest low res date")
						continue
					}

					if lastLowresDate.Time.IsZero() {
						// we have no date for this app, so we need to fetch all data
					} else {
					}

					lastHighresDate, err := u.querier.LastHighResResourceUtilizationDateForApp(ctx, gensql.LastHighResResourceUtilizationDateForAppParams{
						Env:          env,
						ResourceType: gensql.ResourceType(resourceType),
						Team:         team.Slug,
						App:          app.Name,
					})
					if err != nil {
						log.WithError(err).Errorf("unable to fetch latest high res date")
						continue
					}

					if lastHighresDate.Time.IsZero() {
						// we have no date for this app, so we need to fetch all data
					} else {
					}

					// low res
					// high res
					appsUpdated++
				}
			}
		}
	}

	u.log.WithFields(logrus.Fields{
		"duration":    time.Since(start),
		"appsUpdated": appsUpdated,
	}).Debug("resource usage update run finished")

	return nil
}

func previous5minBoundary(t time.Time) time.Time {
	minutesSinceMidnight := t.Hour()*60 + t.Minute()
	remainder := minutesSinceMidnight % 5
	roundedMinutes := minutesSinceMidnight - remainder
	roundedTime := time.Date(t.Year(), t.Month(), t.Day(), roundedMinutes/60, roundedMinutes%60, 0, 0, t.Location())
	return roundedTime
}

func previousHourBoundary(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}


// runResourceUsageUpdater will update resource usage data hourly. This function will block until the context is
// cancelled, so it should be run in a goroutine.
func runResourceUsageUpdater(ctx context.Context, k8sClient *k8s.Client, resourceUsageClient resourceusage.Client, teamsClient *teams.Client, log logrus.FieldLogger) error {
	ticker := time.NewTicker(1 * time.Second) // initial run
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Reset(resourceUpdateSchedule) // regular schedule

			appsUpdated := 0
			start := time.Now()

			// fetch all teams

		}
	}
}


*/
