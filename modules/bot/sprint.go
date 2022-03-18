package tgbot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tg "github.com/nixys/nxs-go-telegram"
)

func sprintState(t *tg.Telegram) (tg.StateHandlerRes, error) {

	buttons := [][]tg.Button{}

	bCtx, b := t.UsrCtxGet().(botCtx)
	if b == false {
		return tg.StateHandlerRes{}, fmt.Errorf("can not extract user context in sprint state handler")
	}

	sprintDateRaw, e, err := t.SlotGet("sprint")
	if err != nil {
		return tg.StateHandlerRes{}, err
	}
	if e == false {
		return tg.StateHandlerRes{
			NextState:    tg.SessState("schedule"),
			StickMessage: true,
		}, nil
	}

	sprintDate := sprintDateRaw.(string)

	sprintIssues, err := bCtx.m.SprintIssuesGetByDate(t.UserIDGet(), sprintDate)
	if err != nil {
		return tg.StateHandlerRes{}, err
	}

	curDate, err := userCurDateGet(t.UserIDGet(), bCtx.m)
	if err != nil {
		return tg.StateHandlerRes{}, err
	}

	c, err := time.Parse("2006-01-02", curDate)
	if err != nil {
		return tg.StateHandlerRes{}, err
	}

	buttons = append(buttons, []tg.Button{
		{
			Text:       "⤴️ Go to selected date: " + c.Format("Monday, 02-Jan-06"),
			Identifier: "date:" + curDate,
		},
	})

	for _, si := range sprintIssues {

		text := si.Text

		if si.Done == true {
			text = "✅ " + text
		}

		if si.Goal == true {
			text = "🎯 " + text
		}

		buttons = append(buttons, []tg.Button{
			{
				Text:       text,
				Identifier: "sprintIssue:" + strconv.Itoa(int(si.ID)),
			},
		})
	}

	return tg.StateHandlerRes{
		Message:      fmt.Sprintf("Issues on sprint %s\n\nEnter text to create new sprint issue", sprintDate),
		Buttons:      buttons,
		StickMessage: true,
	}, nil
}

func sprintMsg(t *tg.Telegram, uc tg.UpdateChain) (tg.MessageHandlerRes, error) {

	bCtx, b := t.UsrCtxGet().(botCtx)
	if b == false {
		return tg.MessageHandlerRes{}, fmt.Errorf("can not extract user context in sprint message handler")
	}

	sprintDateRaw, e, err := t.SlotGet("sprint")
	if err != nil {
		return tg.MessageHandlerRes{}, err
	}
	if e == false {
		return tg.MessageHandlerRes{
			NextState: tg.SessState("schedule"),
		}, nil
	}

	sprintDate := sprintDateRaw.(string)

	sprintIssues, err := bCtx.m.SprintIssuesGetByDate(t.UserIDGet(), sprintDate)
	if err != nil {
		return tg.MessageHandlerRes{}, err
	}

	sprintGoal := false
	if len(sprintIssues) == 0 {
		sprintGoal = true
	}

	if _, err := bCtx.m.SprintIssueAdd(t.UserIDGet(), sprintDate, sprintGoal, strings.Join(uc.MessageTextGet(), "; ")); err != nil {
		return tg.MessageHandlerRes{}, err
	}

	return tg.MessageHandlerRes{
		NextState: tg.SessState("sprint"),
	}, nil
}

func sprintCallback(t *tg.Telegram, uc tg.UpdateChain, identifier string) (tg.CallbackHandlerRes, error) {

	var r tg.CallbackHandlerRes

	action, value, err := buttonIdentifierParse(identifier)
	if err != nil {
		return tg.CallbackHandlerRes{}, err
	}

	switch action {
	case "date":
		r.NextState = tg.SessState("schedule")
	case "sprintIssue":

		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return tg.CallbackHandlerRes{}, err
		}

		if err := t.SlotSave("sprintIssueID", id); err != nil {
			return tg.CallbackHandlerRes{}, err
		}

		r.NextState = tg.SessState("sprintIssueSettings")
	}

	return r, nil
}