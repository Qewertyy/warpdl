package service

import (
	"encoding/json"

	"github.com/warpdl/warpdl/internal/server"
	"github.com/warpdl/warpdl/pkg/warplib"
)

const UPDATE_DOWNLOAD = "download"

type DownloadMessage struct {
	Url               string          `json:"url"`
	DownloadDirectory string          `json:"download_directory"`
	FileName          string          `json:"file_name"`
	Headers           warplib.Headers `json:"headers,omitempty"`
	ForceParts        bool            `json:"force_parts,omitempty"`
	MaxConnections    int             `json:"max_connections,omitempty"`
	MaxSegments       int             `json:"max_segments,omitempty"`
	ChildHash         string          `json:"child_hash,omitempty"`
	IsHidden          bool            `json:"is_hidden,omitempty"`
	IsChildren        bool            `json:"is_children,omitempty"`
}

type DownloadResponse struct {
	DownloadId        string                `json:"download_id"`
	FileName          string                `json:"file_name"`
	SavePath          string                `json:"save_path"`
	DownloadDirectory string                `json:"download_directory"`
	ContentLength     warplib.ContentLength `json:"content_length"`
	Downloaded        warplib.ContentLength `json:"downloaded,omitempty"`
}

const UPDATE_DOWNLOADING = "downloading"

type DownloadingResponse struct {
	DownloadId string `json:"download_id"`
	Action     string `json:"action"`
	Hash       string `json:"hash"`
	Value      int64  `json:"value,omitempty"`
}

func (s *Api) downloadHandler(sconn *server.SyncConn, pool *server.Pool, body json.RawMessage) (string, any, error) {
	var m DownloadMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return UPDATE_DOWNLOAD, nil, err
	}
	var (
		d   *warplib.Downloader
		err error
	)
	d, err = warplib.NewDownloader(s.client, m.Url, &warplib.DownloaderOpts{
		Headers:           m.Headers,
		ForceParts:        m.ForceParts,
		FileName:          m.FileName,
		DownloadDirectory: m.DownloadDirectory,
		MaxConnections:    m.MaxConnections,
		MaxSegments:       m.MaxSegments,
		Handlers: &warplib.Handlers{
			ErrorHandler: func(_ string, err error) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.InitError(err))
				pool.WriteError(uid, server.ErrorTypeCritical, err.Error())
				pool.StopDownload(uid)
				d.Stop()
			},
			DownloadProgressHandler: func(hash string, nread int) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "download_progress",
					Value:      int64(nread),
					Hash:       hash,
				}))
			},
			DownloadCompleteHandler: func(hash string, tread int64) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "download_complete",
					Value:      tread,
					Hash:       hash,
				}))
			},
			DownloadStoppedHandler: func() {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "download_stopped",
				}))
			},
			CompileStartHandler: func(hash string) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_start",
					Hash:       hash,
				}))
			},
			CompileProgressHandler: func(hash string, nread int) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_progress",
					Value:      int64(nread),
					Hash:       hash,
				}))
			},
			CompileCompleteHandler: func(hash string, tread int64) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(UPDATE_DOWNLOADING, &DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_complete",
					Value:      tread,
					Hash:       hash,
				}))
			},
		},
	})
	if err != nil {
		return UPDATE_DOWNLOAD, nil, err
	}
	pool.AddDownload(d.GetHash(), sconn)
	err = s.manager.AddDownload(d, &warplib.AddDownloadOpts{
		ChildHash:        m.ChildHash,
		IsHidden:         m.IsHidden,
		IsChildren:       m.IsChildren,
		AbsoluteLocation: d.GetDownloadDirectory(),
	})
	if err != nil {
		return UPDATE_DOWNLOAD, nil, err
	}
	go d.Start()
	return UPDATE_DOWNLOAD, &DownloadResponse{
		ContentLength:     d.GetContentLength(),
		DownloadId:        d.GetHash(),
		FileName:          d.GetFileName(),
		SavePath:          d.GetSavePath(),
		DownloadDirectory: d.GetDownloadDirectory(),
	}, nil
}