package reperibili

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/axamon/cripta"
)

//Reperibile è la variabile con i dati personali dei reperibili
type Reperibile struct {
	Nome         string
	Cognome      string
	Cellulare    string
	Assegnazione Assegnazione
}

//Assegnazione è la variabile con i dati relativi alla ruota di reperibilità
type Assegnazione struct {
	Piattaforma string
	Giorno      string
	Gruppo      string
}

var t = time.Now()

//limite delle 7 fino alle 7 del mattino seguente il reperibile che viene visualizzato è quello del giorno prima
var limite7 = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())

var ieri = time.Now().Add(-24 * time.Hour).Format("20060102")
var oggi = time.Now().Format("20060102")
var domani = time.Now().Add(24 * time.Hour).Format("20060102")

var filecsv = flag.String("f", "reperibilita.csv", "Percorso del file csv per la reperibilità")
var piattaforma = flag.String("p", "CDN", "La piattaforma di cui desideri ricavare il reperibile")

var contatti []Reperibile

//salva in contatti tutte le informazioni disponibili ora
//var contatti = caricareperibili()

//Reperibiliperpiattaforma2 ti da le info
func Reperibiliperpiattaforma2(piatta, file string) (contatto Reperibile, err error) {
	var limite7 = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())

	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println("errore", err.Error())
	}
	defer csvFile.Close()
	reader := csv.NewReader(bufio.NewReader(csvFile))
	var contatti []Reperibile
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		contatti = append(contatti, Reperibile{

			Nome:      line[3],
			Cognome:   line[4],
			Cellulare: line[5],
			Assegnazione: Assegnazione{
				Giorno:      line[0],
				Piattaforma: line[1],
				Gruppo:      line[2],
			},
		})
	}
	//var reperibili []Reperibile

	switch piatta {
	case "CDN", "TIC":
		for _, contatto := range contatti {
			if contatto.Assegnazione.Giorno == oggi && contatto.Assegnazione.Piattaforma == piatta {
				return contatto, nil
			}
		}

	default:
		for _, contatto := range contatti {
			if t.Before(limite7) {
				//Non sono ancora le 7 di mattina quindi bisogna chiamare il reperibile di ieri
				if contatto.Assegnazione.Giorno == ieri && contatto.Assegnazione.Piattaforma == piatta {
					return contatto, nil
				}
			}
			if t.After(limite7) {
				if contatto.Assegnazione.Giorno == oggi && contatto.Assegnazione.Piattaforma == piatta {
					return contatto, nil
				}
			}

		}

	}
	return contatto, fmt.Errorf("%s", "Nessun reperibile trovato")
}

//Verificacellulare risponde ok se il numero inzia con +3 e si compone di 10 cifre
func Verificacellulare(CELLULARE string) (ok bool) {

	re := regexp.MustCompile(`^\+3[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]$`)
	return re.MatchString(CELLULARE)

}

//Inseriscireperibile inserisce una nuova reperibilità
func Inseriscireperibile(GIORNO, PIATTAFORMA, GRUPPO, NOME, COGNOME, CELLULARE string) (ok bool) {

	GIORNOINT, err := strconv.Atoi(GIORNO)
	if err != nil {
		log.Fatal("Inserito un giorno non nel formato YYYYMMGG")
	}
	oggiint, _ := strconv.Atoi(oggi)
	if GIORNOINT < oggiint {
		log.Fatal("vabbè mo mettemo le reperibilità nel passato")
	}

	if Verificacellulare(CELLULARE) == false {
		log.Fatal("numero di cellulare non supportato, deve essere del tipo +39xxxxxxxxxx")
	}

	value := []string{GIORNO, PIATTAFORMA, GRUPPO, NOME, COGNOME, CELLULARE}

	file, err := os.OpenFile(*filecsv, os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write(value)
	if err != nil {
		log.Fatal(err)
		return false
	}
	writer.Flush()
	return true
}

func recuperavariabile(variabile string) (result string, err error) {
	if result, ok := os.LookupEnv(variabile); ok && len(result) != 0 {
		return result, nil
	}
	return "", fmt.Errorf("la variabile %s non esiste o è vuota", variabile)
}

func recuperavariabilecifrata(variabile, passwd string) (result string, err error) {
	if str, ok := os.LookupEnv(variabile); ok && len(str) != 0 {
		result = cripta.Decifra(str, passwd)
		return result, nil
	}
	return "", fmt.Errorf("la variabile %s non esiste o è vuota", variabile)
}

//Chiamareperibile e comunica il problema
func Chiamareperibile(TO, NOME, COGNOME string) (sid string, err error) {

	twilionumber, err := recuperavariabile("TWILIONUMBER")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	// Let's set some initial default variables

	//Recupera l'accountsid di Twilio dallla variabile d'ambiente
	accountSid, err := recuperavariabile("TWILIOACCOUNTSID")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	//Recupera il token supersegreto dalla variabile d'ambiente
	authToken, err := recuperavariabile("TWILIOAUTHTOKEN")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Calls.json"

	v := url.Values{}
	v.Set("To", TO)
	v.Set("From", twilionumber)
	v.Set("Url", "https://handler.twilio.com/twiml/EH5cef42aa1454fc2326780c8f08c6d568?NOME="+NOME+"&COGNOME="+COGNOME)
	rb := *strings.NewReader(v.Encode())

	client := &http.Client{}

	req, err := http.NewRequest("POST", urlStr, &rb)
	if err != nil {
		fmt.Fprintln(os.Stdout, "OH noooo! Qualcosa è andata storta nel creare la richiesta", err)
	}
	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stdout, "OH noooo! Qualcosa è andata storta nell'inviare la richiesta", err.Error())
	}
	defer resp.Body.Close()
	// make request
	var data map[string]interface{}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {

		bodyBytes, errb := ioutil.ReadAll(resp.Body)
		if errb != nil {
			fmt.Fprintln(os.Stdout, errb.Error())
		}
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			return "", err
		}
	}
	//fmt.Println(data) //debug

	//se la mappa contiene un valore usalo se
	if val, ok := data["sid"]; ok {
		sid = val.(string)
		return sid, nil
	}

	return "", fmt.Errorf("Sid non presente, problemi di auteticazione forse")

}
