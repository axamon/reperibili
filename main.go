package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
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

//salva in contatti tutte le informazioni disponibili ora
var contatti = caricareperibili()

func main() {
	flag.Parse()

	rep, err := Reperibiliperpiattaforma(*piattaforma)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rep.Cellulare)

	ok := Inseriscireperibile("20180515", "PS", "GRUPPO2", "Monica", "Cognome", "+395434520")
	if ok == true {
		fmt.Println("ok")
	}

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
		log.Fatal("vabbè mo mettemo le repribilità nel passato")
	}

	if Verificacellulare(CELLULARE) == false {
		log.Fatal("numero di cellulare non supportato")
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

//Reperibiliperpiattaforma restituisce il reperibile attuale per la piattaforma passata come parametro
func Reperibiliperpiattaforma(piattaforma string) (contatto Reperibile, err error) {

	for _, reperibile := range mostrareperibili() {
		if reperibile.Assegnazione.Piattaforma == piattaforma {
			contatto := reperibile
			return contatto, nil
		}
	}

	return contatto, fmt.Errorf("%s", "Nessun reperibile trovato")
}

func mostrareperibili() (reperibili []Reperibile) {

	//TODO per ogni piattaforma mostra il reperibile attuale
	for _, contatto := range contatti {

		if contatto.Assegnazione.Piattaforma != "CDN" {
			if t.Before(limite7) {
				if contatto.Assegnazione.Giorno == ieri {

					fmt.Println(contatto.Assegnazione.Piattaforma, contatto.Cognome, contatto.Cellulare)

				}
			} else if contatto.Assegnazione.Giorno == oggi {

				reperibili = append(reperibili, contatto)
				//fmt.Println(contatto.Assegnazione.Piattaforma, contatto.Cognome, contatto.Cellulare)

			}

		}
		if contatto.Assegnazione.Giorno == oggi && contatto.Assegnazione.Piattaforma == "CDN" {

			reperibili = append(reperibili, contatto)
			//fmt.Println(contatto.Assegnazione.Piattaforma, contatto.Cognome, contatto.Cellulare)

		}
	}
	return reperibili
}

func caricareperibili() (contatti []Reperibile) {
	csvFile, _ := os.Open(*filecsv)
	defer csvFile.Close()
	reader := csv.NewReader(bufio.NewReader(csvFile))
	//var contatti []Reperibile
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
	//fmt.Println(contatti)
	return
}
