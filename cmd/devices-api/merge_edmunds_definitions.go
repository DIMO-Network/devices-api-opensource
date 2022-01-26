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

func mergeEdmundsDefinitions(ctx context.Context, logger *zerolog.Logger, settings *config.Settings, pdb database.DbStore) error {
	// get non edmunds dd's
	existingDDs, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.NEQ(null.StringFrom(edmundsSource))).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return err
	}
	// loop through non ed dd's, lookup existing match in edmunds dd's,
	for _, dd := range existingDDs {
		// if find match exact match, assume all good and merge
		edmundsDD, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Source.EQ(null.StringFrom(edmundsSource)),
			qm.And("year = ?", dd.Year),
			qm.And("make ilike ?", dd.Make),
			qm.And("model ilike ?", dd.Model)).One(ctx, pdb.DBS().Writer)
		if errors.Is(err, sql.ErrNoRows) {
			// uh oh spaghetti-oh
			fmt.Printf("No Exact Match found for %s\n", printMMY(dd, Red))
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
				stop := askForConfirmation(fmt.Sprintf(" %s No Make and Year matches found in edmunds for: %d %s. Stop to review? %s", Red, dd.Year, dd.Make, Reset))
				if stop {
					os.Exit(0)
				} else {
					continue
				}
			}
			// filter some of the Make and Year matches.
			var modelFirstLetterMatches []*models.DeviceDefinition
			linq.From(edmundsModelYearMatches).Where(func(e interface{}) bool {
				return dd.Model[:1] == e.(*models.DeviceDefinition).Model[:1]
			}).ToSlice(&modelFirstLetterMatches)
			fmt.Printf("Found the following edmunds matches (partial on Model) for: %s:\n", printMMY(dd, Red))
			for i, match := range modelFirstLetterMatches {
				fmt.Printf("%d: %d %s %s\n", i, match.Year, match.Make, match.Model)
			}
			chosenDDToMerge := &models.DeviceDefinition{}

			indexSelection := askForNumberEntry("Choose one from above", len(modelFirstLetterMatches)-1)
			if indexSelection == -2 {
				// prompt to delete
				if hasUserDevices {
					fmt.Println("this DD has userDevices attached so can't be deleted through this tool")
					continue
				} else {
					del := askForConfirmation(fmt.Sprintf("Confirm: %s has no exact edmunds match and no userDevices. Delete? (n to see more options)", printMMY(dd, Red)))
					if del {
						_, err = dd.Delete(ctx, pdb.DBS().Writer)
						if err != nil {
							return errors.Wrapf(err, "error deleting device_definition %s", dd.ID)
						}
						fmt.Println("successfully deleted")
						continue
					}
				}
			}
			if indexSelection == -1 {
				fmt.Printf("ok then, let me show you more edmunds options for %s:\n", printMMY(dd, Red))
				for i, match := range edmundsModelYearMatches {
					fmt.Printf("%d: %d %s %s\n", i, match.Year, match.Make, match.Model)
				}
				indexSelection = askForNumberEntry("Choose one from above", len(edmundsModelYearMatches)-1)
				if indexSelection == -1 {
					// prompt to delete
					del := askForConfirmation(fmt.Sprintf("Ok, no selection then. Would you like to delete Device Def: %s? Delete?", printMMY(dd, Red)))
					if del {
						_, err = dd.Delete(ctx, pdb.DBS().Writer)
						if err != nil {
							return errors.Wrapf(err, "error deleting device_definition %s", dd.ID)
						}
						fmt.Println("successfully deleted")
						continue
					}
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
		fmt.Printf("Found exact match with: %s. ", printMMY(edmundsDD, Green))
		err = mergeMatchingDefinitions(ctx, edmundsDD, dd, pdb)
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeMatchingDefinitions copies make,model,source to existing DD to make it same as edmunds one, moves styles over to existing dd, then deletes the original edmunds one
func mergeMatchingDefinitions(ctx context.Context, edmundsDD, existingDD *models.DeviceDefinition, pdb database.DbStore) error {
	existingDD.Make = edmundsDD.Make
	existingDD.Model = edmundsDD.Model
	existingDD.Source = null.StringFrom(edmundsSource)
	existingDD.ExternalID = edmundsDD.ExternalID
	existingDD.Verified = true
	_, err := existingDD.Update(ctx, pdb.DBS().Writer, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "error updating device_definition with edmunds data")
	}
	// move styles
	edmundsStyles, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(edmundsDD.ID)).All(ctx, pdb.DBS().Writer)
	if err != nil {
		return errors.Wrap(err, "error looking for existing styles")
	}
	stylesUpdatedCnt, err := edmundsStyles.UpdateAll(ctx, pdb.DBS().Writer, models.M{"device_definition_id": existingDD.ID})
	if err != nil {
		return errors.Wrap(err, "error updating all the device_styles to the existing dd_id")
	}
	fmt.Printf("Successfuly updated existing DD with edmunds one, and updated %d device_styles. %s \n",
		stylesUpdatedCnt, printMMY(existingDD, Green))
	_, err = edmundsDD.Delete(ctx, pdb.DBS().Writer)
	if err != nil {
		return errors.Wrapf(err, "error deleting copied edmunds device_definition: %s", edmundsDD.ID)
	}
	return nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error: %v", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// askForNumberEntry asks for a number entry from a range, with max being the highes value. can also take 'n' if no selection desired.
func askForNumberEntry(s string, max int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [0-%d]: (n=none | d=delete): ", s, max)

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

func printMMY(definition *models.DeviceDefinition, color string) string {
	return fmt.Sprintf("%s%d %s %s%s", color, definition.Year, definition.Make, definition.Model, Reset)
}
