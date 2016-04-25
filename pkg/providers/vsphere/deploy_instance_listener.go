package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/layer-x/layerx-commons/lxerrors"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	vsphereInstanceListenerUrl = "https://s3.amazonaws.com/unik-instance-listener/vsphere-instancelistener-base.vmdk"
	vsphereInstanceListenerVmdk = "instancelistener-base.vmdk"
)

func (p *VsphereProvider) DeployInstanceListener() error {
	logrus.Infof("checking if instance listener base vmdk already exists on vsphere datastore")
	c := p.getClient()
	files, err := c.Ls("unik")
	if err != nil {
		return lxerrors.New("lsing on folder 'unik'", err)
	}
	alreadyUploaded := false
	for _, file := range files {
		if strings.Contains(file, vsphereInstanceListenerVmdk) {
			alreadyUploaded = true
			break
		}
	}
	if !alreadyUploaded {
		if _, err := os.Stat(vsphereInstanceListenerVmdk); err != nil {
			logrus.Infof("vbox instance listener vmdk not found, attempting to download from " + vsphereInstanceListenerUrl)
			vmdkFile, err := os.Create(vsphereInstanceListenerVmdk)
			if err != nil {
				return lxerrors.New("creating file for vbox instance listener vmdk", err)
			}
			resp, err := http.Get(vsphereInstanceListenerUrl)
			if err != nil {
				return lxerrors.New("contacting "+ vsphereInstanceListenerUrl, err)
			}
			defer resp.Body.Close()
			n, err := io.Copy(vmdkFile, resp.Body)
			if err != nil {
				return lxerrors.New("copying response to file", err)
			}
			logrus.Infof("%v bytes written", n)
		}
		logrus.Infof("uploading " + vsphereInstanceListenerVmdk)
		if err := c.ImportVmdk(vsphereInstanceListenerVmdk, "unik/"+vsphereInstanceListenerVmdk); err != nil {
			return lxerrors.New("copying instance listener vmdk", err)
		}
	}

	logrus.Infof("deploying virtualbox instance listener")
	if err := c.CreateVm(VsphereInstanceListener, 512); err != nil {
		return lxerrors.New("creating vm", err)
	}
	if err := c.AttachDisk(VsphereInstanceListener, "unik/"+vsphereInstanceListenerVmdk, 0); err != nil {
		return lxerrors.New("attaching disk to vm", err)
	}
	if err := c.PowerOnVm(VsphereInstanceListener); err != nil {
		return lxerrors.New("powering on vm", err)
	}
	return nil
}