package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "net/http"
    "log"
    "github.com/ant0ine/go-json-rest/rest"
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
    gorm.Model
    Players []Player `json:"players" gorm:"many2many:tournament_players;"`
}

type Player struct {
    gorm.Model
    SlackName string
    Name      string
}

func main() {
    i := Impl{}
    i.InitDB()
    i.InitSchema()
    defer i.DB.Close()


    api := rest.NewApi()
    api.Use(rest.DefaultDevStack...)
    router, err := rest.MakeRouter(
        rest.Get("/players", i.GetAllPlayers),
        rest.Post("/players", i.PostPlayer),
    )
    if err != nil {
        log.Fatal(err)
    }
    api.SetApp(router)
    log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))

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
