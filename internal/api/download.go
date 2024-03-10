package api

import (
	"encoding/json"

	"github.com/warpdl/warpdl/common"
	"github.com/warpdl/warpdl/internal/server"
	"github.com/warpdl/warpdl/pkg/warplib"
)

func (s *Api) downloadHandler(sconn *server.SyncConn, pool *server.Pool, body json.RawMessage) (string, any, error) {
	var m common.DownloadParams
	if err := json.Unmarshal(body, &m); err != nil {
		return common.UPDATE_DOWNLOAD, nil, err
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
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "download_progress",
					Value:      int64(nread),
					Hash:       hash,
				}))
			},
			DownloadCompleteHandler: func(hash string, tread int64) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "download_complete",
					Value:      tread,
					Hash:       hash,
				}))
			},
			DownloadStoppedHandler: func() {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "download_stopped",
				}))
			},
			CompileStartHandler: func(hash string) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_start",
					Hash:       hash,
				}))
			},
			CompileProgressHandler: func(hash string, nread int) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_progress",
					Value:      int64(nread),
					Hash:       hash,
				}))
			},
			CompileCompleteHandler: func(hash string, tread int64) {
				uid := d.GetHash()
				pool.Broadcast(uid, server.MakeResult(common.UPDATE_DOWNLOADING, &common.DownloadingResponse{
					DownloadId: uid,
					Action:     "compile_complete",
					Value:      tread,
					Hash:       hash,
				}))
			},
		},
	})
	if err != nil {
		return common.UPDATE_DOWNLOAD, nil, err
	}
	pool.AddDownload(d.GetHash(), sconn)
	err = s.manager.AddDownload(d, &warplib.AddDownloadOpts{
		ChildHash:        m.ChildHash,
		IsHidden:         m.IsHidden,
		IsChildren:       m.IsChildren,
		AbsoluteLocation: d.GetDownloadDirectory(),
	})
	if err != nil {
		return common.UPDATE_DOWNLOAD, nil, err
	}
	// todo: handle download start error
	go d.Start()
	return common.UPDATE_DOWNLOAD, &common.DownloadResponse{
		ContentLength:     d.GetContentLength(),
		DownloadId:        d.GetHash(),
		FileName:          d.GetFileName(),
		SavePath:          d.GetSavePath(),
		DownloadDirectory: d.GetDownloadDirectory(),
	}, nil
}
