# Code-Challenge Golang

## Aufgabe

Erstellen Sie einen REST-Service, der Endpoints zum Verwalten einer
Schadenfallliste bereitstellt. Die einzelnen Schadenfälle sollen dabei
mindestens folgende Eigenschaften besitzen:

* Art des Schadens (max. 100 Zeichen)
* Datum des Eintrags (DD.MM.YYYY)
* Es sollen max. 10 Einträge gespeichert werden dürfen
* Folgende Aktionen sollen über den REST-Service möglich sein:
  * Anlegen eines Eintrags
  * Anzeige aller Einträge
  * Update eines Eintrags
  * Löschen eines Eintrags

## Hinweise

Verwenden Sie zur Umsetzung der Aufgabe folgende Technologien:

* Die Standardbibliothek `net/http` inklusive `http.ServeMux` für das Routing
  (kein externes Web-Framework)
* Swagger-UI zur Visualisierung der APIs
* Für Linting
    ````shell
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
    golangci-lint run
    ````
* Für Testing
    ```shell
     go test ./src/...
     go test -cover ./src/...
    ```

Rahmenbedingungen:

* Es ist ausreichend, wenn der Service alle Werte „in Memory“ zur Laufzeit
  speichert.
* Für alle individuell erstellten Typen sollen Unittests geschrieben werden.
  Die REST-Endpoints sollen über `net/http/httptest` getestet werden.
* Für das Anlegen eines Eintrags muss keine GUI gebaut werden. Dies kann via
  Swagger-UI oder Postman (o. ä.) erfolgen.
* Der REST-Service soll ausschließlich mit JSON arbeiten.

---
