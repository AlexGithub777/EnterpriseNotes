package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/icza/session"
)

type pwrData struct {
	Username string
	Costs    []Cost
}

func (a *App) listHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)

	//get the current username
	sess := session.Get(r)
	user := "[guest]"

	if sess != nil {
		user = sess.CAttr("username").(string)
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}

	rows, err := a.db.Query("SELECT * FROM cost ORDER by id")

	a.checkInternalServerError(err, w)
	var funcMap = template.FuncMap{
		"multiplication": func(n int, f int) int {
			return n * f
		},
		"addOne": func(n int) int {
			return n + 1
		},
	}

	data := pwrData{}
	data.Username = user

	var cost Cost
	for rows.Next() {
		err = rows.Scan(&cost.Id, &cost.ElectricAmount,
			&cost.ElectricPrice, &cost.WaterAmount, &cost.WaterPrice, &cost.CheckedDate)
		a.checkInternalServerError(err, w)
		data.Costs = append(data.Costs, cost)
	}
	t, err := template.New("list.html").Funcs(funcMap).ParseFiles("tmpl/list.html")
	a.checkInternalServerError(err, w)
	err = t.Execute(w, data)
	a.checkInternalServerError(err, w)

}

func (a *App) createHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var cost Cost
	cost.ElectricAmount, _ = strconv.Atoi(r.FormValue("ElectricAmount"))
	cost.ElectricPrice, _ = strconv.Atoi(r.FormValue("ElectricPrice"))
	cost.WaterAmount, _ = strconv.Atoi(r.FormValue("WaterAmount"))
	cost.WaterPrice, _ = strconv.Atoi(r.FormValue("WaterPrice"))
	cost.CheckedDate = r.FormValue("CheckedDate")

	// Save to database
	stmt, err := a.db.Prepare(`
		INSERT INTO cost(electric_amount, electric_price, water_amount, water_price, checked_date)
		VALUES($1, $2, $3, $4, $5)
	`)
	if err != nil {
		fmt.Println("Prepare query error")
		panic(err)
	}
	_, err = stmt.Exec(cost.ElectricAmount, cost.ElectricPrice,
		cost.WaterAmount, cost.WaterPrice, cost.CheckedDate)
	if err != nil {
		fmt.Println("Execute query error")
		panic(err)
	}
	http.Redirect(w, r, "/", 301)
}

func (a *App) updateHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var cost Cost
	cost.Id, _ = strconv.Atoi(r.FormValue("Id"))
	cost.ElectricAmount, _ = strconv.Atoi(r.FormValue("ElectricAmount"))
	cost.ElectricPrice, _ = strconv.Atoi(r.FormValue("ElectricPrice"))
	cost.WaterAmount, _ = strconv.Atoi(r.FormValue("WaterAmount"))
	cost.WaterPrice, _ = strconv.Atoi(r.FormValue("WaterPrice"))
	cost.CheckedDate = r.FormValue("CheckedDate")
	stmt, err := a.db.Prepare(`
		UPDATE cost SET electric_amount=$1, electric_price=$2, water_amount=$3, water_price=$4, checked_date=$5
		WHERE id=$6
	`)
	a.checkInternalServerError(err, w)
	res, err := stmt.Exec(cost.ElectricAmount, cost.ElectricPrice,
		cost.WaterAmount, cost.WaterPrice, cost.CheckedDate, cost.Id)
	a.checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	a.checkInternalServerError(err, w)
	http.Redirect(w, r, "/", 301)

}

func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var costId, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	stmt, err := a.db.Prepare("DELETE FROM cost WHERE id=$1")
	a.checkInternalServerError(err, w)
	res, err := stmt.Exec(costId)
	a.checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	a.checkInternalServerError(err, w)
	http.Redirect(w, r, "/", 301)

}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	http.Redirect(w, r, "/list", 301)
}
