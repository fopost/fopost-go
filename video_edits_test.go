package fopost

import (
	"context"
	"testing"
)

const videoEditJSON = `{"id":"ve_1","status":"succeeded","progress":1,"source_media_id":"med_1","operations":[{"op":"trim","params":{"start":1,"end":5}}],"result_media_id":"med_2","result_media":{"id":"med_2","preview_url":"/v1/media/med_2/file","mime_type":"video/mp4"},"extra_outputs":[{"media_id":"med_3","label":"Clip 1","mime":"video/mp4","preview_url":null}],"result":{"clips":1},"error":null,"created_at":"2026-11-20T10:00:00.000Z","updated_at":"2026-11-20T10:01:00.000Z","completed_at":"2026-11-20T10:01:00.000Z"}`

func TestVideoEditsOps(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":[{"name":"subtitles","available":true,"uses_ai":true}]}`)
	ops, err := client.VideoEdits.Ops(context.Background())
	if err != nil {
		t.Fatalf("VideoEdits.Ops: %v", err)
	}
	if got.method != "GET" || got.path != "/video-edits/ops" {
		t.Fatalf("request = %+v", got)
	}
	if len(ops) != 1 || ops[0].Name != "subtitles" || !ops[0].Available || !ops[0].UsesAI {
		t.Fatalf("ops = %+v", ops)
	}
}

func TestVideoEditsListSendsFilters(t *testing.T) {
	client, got := captureClient(t, 200, `{"data":[`+videoEditJSON+`]}`)
	jobs, err := client.VideoEdits.List(context.Background(), &ListVideoEditsParams{WorkspaceID: "ws_1", MediaID: "med_1", Limit: 5})
	if err != nil {
		t.Fatalf("VideoEdits.List: %v", err)
	}
	if got.method != "GET" || got.path != "/video-edits" || got.query != "limit=5&media_id=med_1&workspace_id=ws_1" {
		t.Fatalf("request = %+v", got)
	}
	if len(jobs) != 1 || jobs[0].Status != VideoEditSucceeded || jobs[0].Operations[0].Op != "trim" {
		t.Fatalf("jobs = %+v", jobs)
	}
}

func TestVideoEditsCreateSendsOperations(t *testing.T) {
	client, got := captureClient(t, 201, `{"id":"ve_1","status":"queued","progress":0,"source_media_id":"med_1","operations":[{"op":"trim","params":{"start":1,"end":5}}],"result_media_id":null,"result_media":null,"extra_outputs":[],"result":null,"error":null,"created_at":"2026-11-20T10:00:00.000Z","updated_at":"2026-11-20T10:00:00.000Z","completed_at":null}`)
	job, err := client.VideoEdits.Create(context.Background(), &CreateVideoEditRequest{
		WorkspaceID: "ws_1",
		MediaID:     "med_1",
		Operations: []VideoEditOperation{
			{Op: "trim", Params: map[string]any{"start": 1, "end": 5}},
			{Op: "denoise"},
		},
	})
	if err != nil {
		t.Fatalf("VideoEdits.Create: %v", err)
	}
	if got.method != "POST" || got.path != "/video-edits" || got.body["media_id"] != "med_1" || got.body["workspace_id"] != "ws_1" {
		t.Fatalf("request = %+v", got)
	}
	ops, _ := got.body["operations"].([]any)
	first, _ := ops[0].(map[string]any)
	second, _ := ops[1].(map[string]any)
	if len(ops) != 2 || first["op"] != "trim" || first["params"].(map[string]any)["end"] != float64(5) {
		t.Fatalf("operations = %+v", ops)
	}
	if _, ok := second["params"]; ok {
		t.Fatalf("empty params should be omitted: %+v", second)
	}
	if job.Status != VideoEditQueued || job.ResultMedia != nil || job.ResultMediaID != nil || job.CompletedAt != nil {
		t.Fatalf("job = %+v", job)
	}
}

func TestVideoEditsCreateValidates(t *testing.T) {
	client, _ := captureClient(t, 201, `{}`)
	if _, err := client.VideoEdits.Create(context.Background(), &CreateVideoEditRequest{Operations: []VideoEditOperation{{Op: "trim"}}}); err == nil {
		t.Fatal("expected an error without a media id")
	}
	if _, err := client.VideoEdits.Create(context.Background(), &CreateVideoEditRequest{MediaID: "med_1"}); err == nil {
		t.Fatal("expected an error without operations")
	}
}

func TestVideoEditsGetDecodesBareJob(t *testing.T) {
	client, got := captureClient(t, 200, videoEditJSON)
	job, err := client.VideoEdits.Get(context.Background(), "ve_1", "ws_1")
	if err != nil {
		t.Fatalf("VideoEdits.Get: %v", err)
	}
	if got.method != "GET" || got.path != "/video-edits/ve_1" || got.query != "workspace_id=ws_1" {
		t.Fatalf("request = %+v", got)
	}
	if job.ResultMedia == nil || *job.ResultMedia.MimeType != "video/mp4" || len(job.ExtraOutputs) != 1 || job.ExtraOutputs[0].Label != "Clip 1" {
		t.Fatalf("job = %+v", job)
	}
	if job.Result["clips"] != float64(1) || job.Error != nil || job.CompletedAt == nil {
		t.Fatalf("job = %+v", job)
	}
}
