package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session, payload *message.SignalVideo) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.Profile().CanWatch {
		logger.Debug().Msg("not allowed to watch")
		return nil
	}

	// use default first video, if not provided
	if payload.Video == "" {
		videos := h.capture.VideoIDs()
		payload.Video = videos[0]
	}

	offer, err := h.webrtc.CreatePeer(session, payload.Video)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_PROVIDE,
		message.SignalProvide{
			SDP:        offer.SDP,
			ICEServers: h.webrtc.ICEServers(),
			Video:      payload.Video,
		})

	return nil
}

func (h *MessageHandlerCtx) signalRestart(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	peer := session.GetWebRTCPeer()
	if peer == nil {
		logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	offer, err := peer.CreateOffer(true)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_RESTART,
		message.SignalAnswer{
			SDP: offer.SDP,
		})

	return nil
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalAnswer) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	peer := session.GetWebRTCPeer()
	if peer == nil {
		logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalAnswer(payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	peer := session.GetWebRTCPeer()
	if peer == nil {
		logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalCandidate(payload.ICECandidateInit)
}

func (h *MessageHandlerCtx) signalVideo(session types.Session, payload *message.SignalVideo) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	peer := session.GetWebRTCPeer()
	if peer == nil {
		logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	err := peer.SetVideoID(payload.Video)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video: payload.Video,
		})

	return nil
}
