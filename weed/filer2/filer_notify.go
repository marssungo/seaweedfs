package filer2

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/notification"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/util"
)

func (f *Filer) NotifyUpdateEvent(oldEntry, newEntry *Entry, deleteChunks bool) {
	var fullpath string
	if oldEntry != nil {
		fullpath = string(oldEntry.FullPath)
	} else if newEntry != nil {
		fullpath = string(newEntry.FullPath)
	} else {
		return
	}

	// println("fullpath:", fullpath)

	if strings.HasPrefix(fullpath, "/.meta") {
		return
	}

	newParentPath := ""
	if newEntry != nil {
		newParentPath, _ = newEntry.FullPath.DirAndName()
	}
	eventNotification := &filer_pb.EventNotification{
		OldEntry:      oldEntry.ToProtoEntry(),
		NewEntry:      newEntry.ToProtoEntry(),
		DeleteChunks:  deleteChunks,
		NewParentPath: newParentPath,
	}

	if notification.Queue != nil {
		glog.V(3).Infof("notifying entry update %v", fullpath)
		notification.Queue.SendMessage(fullpath, eventNotification)
	}

	f.logMetaEvent(time.Now(), fullpath, eventNotification)

}

func (f *Filer) logMetaEvent(ts time.Time, fullpath string, eventNotification *filer_pb.EventNotification) {

	dir, _ := util.FullPath(fullpath).DirAndName()

	event := &filer_pb.FullEventNotification{
		Directory:         dir,
		EventNotification: eventNotification,
	}
	data, err := proto.Marshal(event)
	if err != nil {
		glog.Errorf("failed to marshal filer_pb.FullEventNotification %+v: %v", event, err)
		return
	}

	f.metaLogBuffer.AddToBuffer(ts, []byte(dir), data)

}

func (f *Filer) logFlushFunc(startTime, stopTime time.Time, buf []byte) {
	targetFile := fmt.Sprintf("/.meta/log/%04d/%02d/%02d/%02d/%02d/%02d-%02d.log",
		startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(),
		startTime.Second(), stopTime.Second())

	if err := f.appendToFile(targetFile, buf); err != nil {
		glog.V(0).Infof("log write failed %s: %v", targetFile, err)
	}
}

