package telegram

const msgHelp = `I can save your login details for various services

To save the data do the following iterations
1) type /set
2) enter the name of the service
3) enter your username and password to enter this service separated by a space
Attention!
Data for identical service names is overwritten

In order to get or delete data, perform the following iterations
1) type /get or /del
2) enter the name of the service`

const msgHello = "Hi there! 👽\n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command 🤔"
	msgNoSaved        = "You have no saved service 😔 \n\n"
	msgSaved          = "Saved! 👌"
	msgDelete         = "Deleted! 🗑"
	msgSetService     = "Enter service name 🖌️"
	msgNoSavedService = "You have no saved data for this service 😔 \n\n" + msgSetService
	msgSetData        = "Enter your username and password separated by a space 🔓"
	msgWrongFormat    = "Invalid input format 🥴\n\n" + msgSetData
)
