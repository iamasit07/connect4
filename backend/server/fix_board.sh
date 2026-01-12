#!/bin/bash
# Fix line 200: add board parameter
sed -i '200s/gs.FinishedAt)/gs.FinishedAt, convertBoardTo Ints(gs.Game.Board))/' game_session.go
# Fix line 284
sed -i '284s/gs.FinishedAt)/gs.FinishedAt, convertBoardToInts(gs.Game.Board))/' game_session.go
# Fix line 381
sed -i '381s/gs.FinishedAt)/gs.FinishedAt, convertBoardToInts(gs.Game.Board))/' game_session.go
# Fix line 474
sed -i '474s/gs.FinishedAt,/gs.FinishedAt, convertBoardToInts(gs.Game.Board),/' game_session.go
# Fix line 533
sed -i '533s/time.Now())/time.Now(), convertBoardToInts(gs.Game.Board))/' game_session.go
