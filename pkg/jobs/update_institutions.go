package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/plaid/plaid-go/plaid"
)

const (
	UpdateInstitutions = "UpdateInstitutions"
)

func (j *jobManagerBase) updateInstitutions(job *work.Job) error {
	span := sentry.StartSpan(context.Background(), "Job", sentry.TransactionName("Update Institutions"))
	defer span.Finish()

	log := j.getLogForJob(job)

	institutions, err := j.plaidClient.GetAllInstitutions(span.Context(), []string{"US"}, plaid.GetInstitutionsOptions{
		Products: []string{
			"transactions",
		},
		IncludeOptionalMetadata: true,
	})
	if err != nil {
		log.WithError(err).Error("failed to retrieve institutions for update")
		return err
	}

	if len(institutions) == 0 {
		log.Warn("no institutions found")
		return nil
	}

	plaidInstitutionIds := make([]string, len(institutions))
	for i, institution := range institutions {
		plaidInstitutionIds[i] = institution.ID
	}

	return j.getJobHelperRepository(job, func(repo repository.JobRepository) error {
		byPlaidId, err := repo.GetInstitutionsByPlaidID(span.Context(), plaidInstitutionIds)
		if err != nil {
			return err
		}

		institutionsToUpdate := make([]*models.Institution, 0)
		institutionsToCreate := make([]*models.Institution, 0)
		for _, institution := range institutions {
			existingInstitution, ok := byPlaidId[institution.ID]
			if !ok {
				var url, color, logo *string
				if institution.URL != "" {
					url = &institution.URL
				}

				if institution.PrimaryColor != "" {
					color = &institution.PrimaryColor
				}

				if institution.Logo != "" {
					logo = &institution.Logo
				}

				institutionsToCreate = append(institutionsToCreate, &models.Institution{
					Name:               institution.Name,
					PlaidInstitutionId: &institution.ID,
					PlaidProducts:      institution.Products,
					URL:                url,
					PrimaryColor:       color,
					Logo:               logo,
				})
			}

			if existingInstitution.Name != institution.Name {
				existingInstitution.Name = institution.Name
				institutionsToUpdate = append(institutionsToUpdate, &existingInstitution)
			}
		}

		if len(institutionsToUpdate) == 0 && len(institutionsToCreate) == 0 {
			log.Infof("no institution changes")
			return nil
		}

		if len(institutionsToUpdate) > 0 {
			log.Infof("updating %d institutions", len(institutionsToUpdate))
			if err = repo.UpdateInstitutions(span.Context(), institutionsToUpdate); err != nil {
				return err
			}
		}

		if len(institutionsToCreate) > 0 {
			log.Infof("creating %d institutions", len(institutionsToCreate))
			if err = repo.CreateInstitutions(span.Context(), institutionsToCreate); err != nil {
				return err
			}
		}

		return nil
	})
}
