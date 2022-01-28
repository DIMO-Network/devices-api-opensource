package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DIMO-INC/devices-api/internal/config"
	"github.com/DIMO-INC/devices-api/internal/database"
	"github.com/DIMO-INC/devices-api/models"
	"github.com/ahmetb/go-linq/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var Red = "\033[31m"
var Reset = "\033[0m"
var Green = "\033[32m"
var Purple = "\033[35m"

func mergeEdmundsDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get non edmunds dd's
	existingDDs, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.NEQ(null.StringFrom(edmundsSource)),
		qm.Or("source not like %cli-ignored")).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return err
	}
	// loop through non ed dd's, lookup existing match in edmunds dd's,
	for _, dd := range existingDDs {
		fmt.Println("--------------------------------------------")
		// if find match exact match, assume all good and merge
		edmundsDD, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.EQ(null.StringFrom(edmundsSource)),
			qm.And("year = ?", dd.Year),
			qm.And("make ilike ?", dd.Make),
			qm.And("model ilike ?", dd.Model)).One(ctx, pdb.DBS().Writer)
		if errors.Is(err, sql.ErrNoRows) {
			// uh oh spaghetti-oh
			fmt.Printf("No Exact Match found for %s\n", printMMY(dd, Red, true))
			// lookup if has userDevices, if not offer to delete
			hasUserDevices, err := models.UserDevices(models.UserDeviceWhere.DeviceDefinitionID.EQ(dd.ID)).Exists(ctx, pdb.DBS().Writer)
			if err != nil {
				return errors.Wrap(err, "error looking up user devices for dd: "+dd.ID)
			}
			// get all make and year matches from edmunds, then lookup by first letter of model and show responses with a picker.
			edmundsModelYearMatches, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.EQ(null.StringFrom(edmundsSource)),
				qm.And("year = ?", dd.Year),
				qm.And("make ilike ?", dd.Make)).All(ctx, pdb.DBS().Writer)
			if err != nil {
				return errors.Wrap(err, "error querying all make and year edmunds matches")
			}
			// if no Make & Year matches here likely means something off in our DB, offer to stop to review
			if len(edmundsModelYearMatches) == 0 {
				del := askForConfirmation(fmt.Sprintf(" %s No Make and Year matches found in edmunds for: %d %s. Delete? Ignore? %s", Red, dd.Year, dd.Make, Reset))
				if del == nil {
					//mark ignored
					markIgnored(dd)
				} else if *del {
					_, err = dd.Delete(ctx, pdb.DBS().Writer)
					if err != nil {
						return errors.Wrapf(err, "error deleting device_definition %s", dd.ID)
					}
					fmt.Println("successfully deleted")
				}
				continue
			}
			// filter some of the Make and Year matches.
			var modelFirstLetterMatches []*models.DeviceDefinition
			linq.From(edmundsModelYearMatches).Where(func(e interface{}) bool {
				return dd.Model[:1] == e.(*models.DeviceDefinition).Model[:1]
			}).ToSlice(&modelFirstLetterMatches)
			fmt.Printf("Found the following edmunds matches (partial on Model) for: %s:\n", printMMY(dd, Red, true))
			for i, match := range modelFirstLetterMatches {
				fmt.Printf("%d: %d %s %s\n", i, match.Year, match.Make, match.Model)
			}
			chosenDDToMerge := &models.DeviceDefinition{} //nolint

			indexSelection := askForNumberEntry("Choose one from above", len(modelFirstLetterMatches)-1)
			if indexSelection == -3 {
				markIgnored(dd)
				continue
			}
			if indexSelection == -2 {
				// prompt to delete
				if hasUserDevices {
					fmt.Println("this DD has userDevices attached so can't be deleted through this tool")
					continue
				} else {
					del := askForConfirmation(fmt.Sprintf("Confirm: %s has no exact edmunds match and no userDevices. Delete? (n to see more options)", printMMY(dd, Red, false)))
					if del == nil {
						markIgnored(dd)
					} else if *del {
						_, err = dd.Delete(ctx, pdb.DBS().Writer)
						if err != nil {
							return errors.Wrapf(err, "error deleting device_definition %s", dd.ID)
						}
						fmt.Println("successfully deleted")
					}
					continue
				}
			}
			if indexSelection == -1 {
				fmt.Printf("ok then, let me show you more edmunds options for %s:\n", printMMY(dd, Red, false))
				for i, match := range edmundsModelYearMatches {
					fmt.Printf("%d: %d %s %s\n", i, match.Year, match.Make, match.Model)
				}
				indexSelection = askForNumberEntry("Choose one from above", len(edmundsModelYearMatches)-1)
				if indexSelection == -3 {
					markIgnored(dd)
					continue
				}
				if indexSelection == -1 {
					// prompt to delete
					del := askForConfirmation(fmt.Sprintf("Ok, no selection then. Would you like to delete Device Def: %s? Delete?", printMMY(dd, Red, true)))
					if del == nil {
						markIgnored(dd)
					} else if *del {
						_, err = dd.Delete(ctx, pdb.DBS().Writer)
						if err != nil {
							return errors.Wrapf(err, "error deleting device_definition %s", dd.ID)
						}
						fmt.Println("successfully deleted" + printMMY(dd, Red, false))
					}
					continue
				} else {
					chosenDDToMerge = edmundsModelYearMatches[indexSelection]
				}
			} else {
				chosenDDToMerge = modelFirstLetterMatches[indexSelection]
			}
			if chosenDDToMerge != nil {
				err = mergeMatchingDefinitions(ctx, chosenDDToMerge, dd, pdb)
				if err != nil {
					return err
				}
			}

			continue
		}
		if err != nil {
			return err
		}
		fmt.Printf("Found exact match with: %s. ", printMMY(edmundsDD, Green, false))
		err = mergeMatchingDefinitions(ctx, edmundsDD, dd, pdb)
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeMatchingDefinitions moves existing DD to the edmunds DD, moving over all related data.
func mergeMatchingDefinitions(ctx context.Context, edmundsDD, existingDD *models.DeviceDefinition, pdb database.DbStore) error {
	tx, err := pdb.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "could not start transaction")
	}
	defer tx.Rollback() //nolint

	// bring over user devices to edmunds DD
	existingDDUserDevices, err := models.UserDevices(models.UserDeviceWhere.DeviceDefinitionID.EQ(existingDD.ID)).All(ctx, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "error getting edmunds userDevices for dd_id %s", edmundsDD.ID)
	}
	if len(existingDDUserDevices) > 0 {
		fmt.Printf("found %d userDevices attached to existing dd_id %s, moving them over to edmunds dd_id %s\n",
			len(existingDDUserDevices), existingDD.ID, edmundsDD.ID)
		_, err = existingDDUserDevices.UpdateAll(ctx, tx, models.M{"device_definition_id": edmundsDD.ID})
		if err != nil {
			return errors.Wrap(err, "error updating user devices dd_id")
		}
	}
	// bring over any integrations to edmunds DD
	existingDDIntegrations, err := models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(existingDD.ID)).All(ctx, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrapf(err, "error getting edmunds integrations for dd_id %s", edmundsDD.ID)
	}
	if len(existingDDIntegrations) > 0 {
		fmt.Printf("found deviceIntegrations attached to existing dd_id %s, moving them over to edmunds dd_id %s\n", existingDD.ID, edmundsDD.ID)
		exists, err := models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(edmundsDD.ID)).Exists(ctx, tx)
		if err != nil {
			return err
		}
		if !exists {
			_, err = existingDDIntegrations.UpdateAll(ctx, tx, models.M{"device_definition_id": edmundsDD.ID})
			if err != nil {
				fmt.Printf("error trying to update deviceIntegrations, most likely because conflict with existing integration, which is OK. err: %v \n", err)
			}
		}
	}
	// copy any useful old data
	if existingDD.ImageURL.Ptr() != nil || existingDD.Metadata.Ptr() != nil {
		edmundsDD.ImageURL = existingDD.ImageURL
		edmundsDD.Metadata = existingDD.Metadata
		_, err = edmundsDD.Update(ctx, tx, boil.Infer())
		if err != nil {
			return errors.Wrap(err, "error updating device_definition with edmunds data")
		}
	}

	fmt.Printf("Successfuly updated edmunds DD with selected existing one. %s\n", printMMY(existingDD, Green, false))
	_, err = existingDD.Delete(ctx, tx)
	if err != nil {
		return errors.Wrapf(err, "error deleting existing device_definition: %s", existingDD.ID)
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "error commiting transaction")
	}
	return nil
}

func askForConfirmation(s string) *bool {
	reader := bufio.NewReader(os.Stdin)
	c := false
	for {
		fmt.Printf("%s [y/n/i]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error: %v", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			c = true
			return &c
		} else if response == "n" || response == "no" {
			return &c
		} else if response == "i" {
			return nil
		}
	}
}

// askForNumberEntry asks for a number entry from a range, with max being the highes value. can also take 'n' if no selection desired.
func askForNumberEntry(s string, max int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [0-%d]: (n=none | d=delete | i=ignore): ", s, max)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "n" {
			return -1
		}
		if response == "d" {
			return -2
		}
		if response == "i" {
			return -3
		}

		number, err := strconv.Atoi(response)
		if err != nil {
			fmt.Println("invalid entry, try again: " + response)
			continue
		}
		if number > max {
			fmt.Println("number to big, try again: " + response)
			continue
		}
		return number
	}
}

func printMMY(definition *models.DeviceDefinition, color string, includeSource bool) string {
	if !includeSource {
		return fmt.Sprintf("%s%d %s %s%s", color, definition.Year, definition.Make, definition.Model, Reset)
	}
	return fmt.Sprintf("%s%d %s %s %s(source: %s)%s",
		color, definition.Year, definition.Make, definition.Model, Purple, definition.Source.String, Reset)
}

func markIgnored(definition *models.DeviceDefinition) {
	if definition.Source.Ptr() == nil {
		definition.Source = null.StringFrom("cli-ignored")
	} else {
		definition.Source = null.StringFrom(fmt.Sprintf("%s, cli-ignored", definition.Source.String))
	}
}
