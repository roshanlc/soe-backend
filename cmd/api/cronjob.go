// This file contains method for cronjob to remove expired tokens
package main

import "log"

// This function will remove expired tokens
func (app *application) expiredTokenRemoval() {

	log.Println("Performing expired token removal..")
	err := app.models.Tokens.RemoveExpiredTokens()

	if err != nil {
		log.Println("Error Occured while removing tokens, ", err)
	}

}
