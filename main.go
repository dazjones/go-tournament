package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
)

type Impl struct {
	DB *gorm.DB
}

func (i *Impl) InitDB() {
	var err error
	i.DB, err = gorm.Open("sqlite3", "tournament.db")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
	i.DB.LogMode(true)
}

func (i *Impl) InitSchema() {
	i.DB.CreateTable(&Tournament{}, &Player{})
	i.DB.Model(&Tournament{}).Related(&Player{}, "Players")
	i.DB.AutoMigrate(&Tournament{}, &Player{})
}

type Tournament struct {
	ID      uint     `gorm:"primary_key"`
	Players []Player `json:"players" gorm:"many2many:tournament_players;"`
}

type Player struct {
	ID        uint   `gorm:"primary_key"`
	SlackName string `json:"slack_name" gorm:"not null;unique"`
	Name      string `json:"name" gorm:"not null"`
}

func main() {
	i := Impl{}
	i.InitDB()
	i.InitSchema()
	defer i.DB.Close()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/status", i.Status),
		rest.Get("/players", i.GetAllPlayers),
		rest.Post("/players", i.PostPlayer),
		rest.Get("/tournaments", i.GetAllTournaments),
		rest.Post("/tournaments", i.PostTournament),
		rest.Get("/", i.Home),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))

}

func (i *Impl) Home(w rest.ResponseWriter, r *rest.Request) {
    w.WriteJson("go-tournament is up and running!")
}

func (i *Impl) GetAllPlayers(w rest.ResponseWriter, r *rest.Request) {
	players := []Player{}
	i.DB.Find(&players)
	w.WriteJson(&players)
}

func (i *Impl) PostPlayer(w rest.ResponseWriter, r *rest.Request) {
	player := Player{}
	if err := r.DecodeJsonPayload(&player); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := i.DB.Save(&player).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&player)
}

func (i *Impl) GetAllTournaments(w rest.ResponseWriter, r *rest.Request) {
	tournaments := []Tournament{}
	i.DB.Find(&tournaments)
	w.WriteJson(&tournaments)
}

func (i *Impl) PostTournament(w rest.ResponseWriter, r *rest.Request) {
	tournament := Tournament{}
	if err := r.DecodeJsonPayload(&tournament); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := i.DB.Save(&tournament).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&tournament)
}

func (i *Impl) Status(w rest.ResponseWriter, r *rest.Request) {
    w.WriteJson(map[string]string{"status": "OK"})
}
