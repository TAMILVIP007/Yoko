package main

import (
 tb "gopkg.in/tucnak/telebot.v2"
 "fmt"
 "context"
 "go.mongodb.org/mongo-driver/mongo/readpref"
)

func lock(m *tb.Message){
 x := db.Ping(context.TODO(), readpref.Primary())
 if x == nil{
  fmt.Println("z")
 }
}
