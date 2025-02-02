package main

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "html/template"
        "log"
        "net/http"
        "strconv"
        "encoding/csv"

        _ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
        "github.com/gorilla/mux"
)

// Risk struct
type Risk struct {
        ItemName          string `json:"item_name"`              
        ID                int    `json:"id"`
        RiskID            string `json:"risk_id"`
        RiskDescription   string `json:"risk_description"`
        Likelihood        int    `json:"likelihood"`
        Impact            int    `json:"impact"`
        RiskScore         int    `json:"risk_score"`
        MitigationActions string `json:"mitigation_actions"`
        ResponsiblePerson string `json:"responsible_person"`
    ResponsibleDepartment string `json:"responsible_department"`
        TargetCompletionDate string `json:"target_completion_date"`
        Status            string `json:"status"`
        Stakeholders      string `json:"stakeholders"`
}

var db *sql.DB
var tmpl *template.Template

func main() {
// Default db connection to local SQLite3
	    var err error
        db, err = sql.Open("sqlite3", "risk_register.db") // Open or create the database file
        if err != nil {
                log.Fatal(err)
        }
        defer db.Close()

               // Create table if it doesn't exist (important for SQLite!)
               _, err = db.Exec(`
               CREATE TABLE IF NOT EXISTS risks (
                       item_name text,
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       risk_id TEXT,
                       risk_description TEXT,
                       likelihood INTEGER,
                       impact INTEGER,
                       risk_score INTEGER,
                       mitigation_actions TEXT,
                       responsible_person TEXT,
                       responsible_department TEXT,
                       target_completion_date TEXT,
                       status TEXT,
                       stakeholders TEXT
               );
       `)
       if err != nil {
               log.Fatal(err)
       }

        tmpl = template.Must(template.ParseGlob("templates/*.html"))

        r := mux.NewRouter()

        r.HandleFunc("/risks", getRisks).Methods("GET")
        r.HandleFunc("/risks/{id}", getRisk).Methods("GET")
        r.HandleFunc("/risks", createRisk).Methods("POST")
        r.HandleFunc("/risks/{id}", updateRisk).Methods("PUT")
        r.HandleFunc("/risks/{id}", deleteRisk).Methods("DELETE")

        r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
        r.HandleFunc("/", homeHandler)
	    r.HandleFunc("/settings", settingsHandler) // New route for settings
        r.HandleFunc("/import", importHandler) // route for import form
        r.HandleFunc("/risks/import", importRisks).Methods("POST")

        fmt.Println("Server started on LOCALHOST port 8080...")
        log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
        err := tmpl.ExecuteTemplate(w, "index.html", nil)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
    err := tmpl.ExecuteTemplate(w, "settings.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func importHandler(w http.ResponseWriter, r *http.Request) {
    err := tmpl.ExecuteTemplate(w, "import.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func getRisks(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT * FROM risks")
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        defer rows.Close()

        var risks []Risk
        for rows.Next() {
                var risk Risk
                err := rows.Scan(&risk.ItemName, &risk.ID, &risk.RiskID, &risk.RiskDescription, &risk.Likelihood, &risk.Impact, &risk.RiskScore, &risk.MitigationActions, &risk.ResponsiblePerson, &risk.ResponsibleDepartment, &risk.TargetCompletionDate, &risk.Status, &risk.Stakeholders)
                if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                }
                risks = append(risks, risk)
        }

        w.Header().Set("Content-Type", "application/json") // Important!
        json.NewEncoder(w).Encode(risks)
}

func getRisk(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid risk ID", http.StatusBadRequest)
        return
    }

    var risk Risk
    err = db.QueryRow("SELECT * FROM risks WHERE id = ?", id).Scan(&risk.ItemName, &risk.ID, &risk.RiskID, &risk.RiskDescription, &risk.Likelihood, &risk.Impact, &risk.RiskScore, &risk.MitigationActions, &risk.ResponsiblePerson, &risk.ResponsibleDepartment, &risk.TargetCompletionDate, &risk.Status, &risk.Stakeholders)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Risk not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json") // Important!
    json.NewEncoder(w).Encode(risk)
}

func createRisk(w http.ResponseWriter, r *http.Request) {
        var risk Risk
        err := json.NewDecoder(r.Body).Decode(&risk)
        if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        result, err := db.Exec("INSERT INTO risks (item_name, risk_id, risk_description, likelihood, impact, risk_score, mitigation_actions, responsible_person, responsible_department, target_completion_date, status, stakeholders) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                risk.RiskID, risk.RiskDescription, risk.Likelihood, risk.Impact, risk.RiskScore, risk.MitigationActions, risk.ResponsiblePerson, risk.ResponsibleDepartment, risk.TargetCompletionDate, risk.Status, risk.Stakeholders)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        id, _ := result.LastInsertId()
        risk.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(risk)
}

func updateRisk(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid risk ID", http.StatusBadRequest)
        return
    }

    var risk Risk
    err = json.NewDecoder(r.Body).Decode(&risk)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE risks SET item_name = ?, risk_id = ?, risk_description = ?, likelihood = ?, impact = ?, risk_score = ?, mitigation_actions = ?, responsible_person = ?, responsible_department = ?, target_completion_date = ?, status = ?, stakeholders = ? WHERE id = ?",
        risk.RiskID, risk.RiskDescription, risk.Likelihood, risk.Impact, risk.RiskScore, risk.MitigationActions, risk.ResponsiblePerson, risk.ResponsibleDepartment, risk.TargetCompletionDate, risk.Status, risk.Stakeholders, id)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(risk)

}

func deleteRisk(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid risk ID", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM risks WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func importRisks(w http.ResponseWriter, r *http.Request) {
    // 1. Get the uploaded file
    file, _, err := r.FormFile("file") // "file" is the name of the file input field in the HTML form
    if err != nil {
        http.Error(w, "Error retrieving the file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // 2. Parse the CSV file
    reader := csv.NewReader(file)
    reader.Comma = ',' // Set the delimiter (comma by default)
    reader.FieldsPerRecord = 13 //Important, adjust if you have more or less fields

    records, err := reader.ReadAll()
    if err != nil {
        http.Error(w, "Error parsing CSV file", http.StatusBadRequest)
        return
    }

    // 3. Validate and insert risks
    var importedRisks []Risk
    for i, record := range records {
        if i == 0 { // Skip header row
            continue
        }

        if len(record) != 13 {
            http.Error(w, fmt.Sprintf("Invalid number of columns on row %d", i+1), http.StatusBadRequest)
            return
        }

        likelihood, err := strconv.Atoi(record[4])
        if err != nil {
            http.Error(w, fmt.Sprintf("Invalid likelihood value on row %d", i+1), http.StatusBadRequest)
            return
        }

        impact, err := strconv.Atoi(record[5])
        if err != nil {
            http.Error(w, fmt.Sprintf("Invalid impact value on row %d", i+1), http.StatusBadRequest)
            return
        }

        riskScore, err := strconv.Atoi(record[6])
        if err != nil {
            http.Error(w, fmt.Sprintf("Invalid risk score value on row %d", i+1), http.StatusBadRequest)
            return
        }

        risk := Risk{
            ItemName:          record[0],
            RiskID:            record[1],
            RiskDescription:   record[2],
            Likelihood:        likelihood,
            Impact:            impact,
            RiskScore:         riskScore,
            MitigationActions: record[7],
            ResponsiblePerson: record[8],
            ResponsibleDepartment: record[9],
            TargetCompletionDate: record[10],
            Status:            record[11],
            Stakeholders:      record[12],
        }

        result, err := db.Exec("INSERT INTO risks (item_name, risk_id, risk_description, likelihood, impact, risk_score, mitigation_actions, responsible_person, responsible_department, target_completion_date, status, stakeholders) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        risk.RiskID, risk.RiskDescription, risk.Likelihood, risk.Impact, risk.RiskScore, risk.MitigationActions, risk.ResponsiblePerson, risk.ResponsibleDepartment, risk.TargetCompletionDate, risk.Status, risk.Stakeholders)

        if err != nil {
            http.Error(w, fmt.Sprintf("Error inserting risk on row %d: %v", i+1, err), http.StatusInternalServerError)
            return
        }
        id, _ := result.LastInsertId()
        risk.ID = int(id)
        importedRisks = append(importedRisks, risk)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(importedRisks)
}

