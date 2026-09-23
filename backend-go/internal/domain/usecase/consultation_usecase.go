package usecase

import (
	"context"
)

type AgoraTokenResponse struct {
	Token       string `json:"token"`
	ChannelName string `json:"channel_name,omitempty"`
	Channel     string `json:"channel,omitempty"`
	UID         uint32 `json:"uid,omitempty"`
	AppID       string `json:"app_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

type ConsultationUseCase interface {
	GetRTCToken(ctx context.Context, channelName string, uid uint32, role string) (*AgoraTokenResponse, error)
	GetRTMToken(ctx context.Context, userAccount string) (*AgoraTokenResponse, error)
}
