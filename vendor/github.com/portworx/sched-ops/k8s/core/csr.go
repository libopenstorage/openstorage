package core

import (
	"context"
	"time"

	certv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// CertificateOps is an interface to perform k8s Certificate operations
type CertificateOps interface {
	CreateCertificateSigningRequests(csr []byte, name string, labs map[string]string, signer string, dur *time.Duration, usages []certv1.KeyUsage) (*certv1.CertificateSigningRequest, error)
	UpdateCertificateSigningRequests(csr []byte, name string, labs map[string]string, signer string, dur *time.Duration, usages []certv1.KeyUsage) (*certv1.CertificateSigningRequest, error)
	ListCertificateSigningRequests(labels map[string]string) (*certv1.CertificateSigningRequestList, error)
	GetCertificateSigningRequest(name string) (*certv1.CertificateSigningRequest, error)
	DeleteCertificateSigningRequests(name string) error
	WatchCertificateSigningRequests(csr *certv1.CertificateSigningRequest, fn WatchFunc) error
	CertificateSigningRequestsUpdateApproval(name string, csr *certv1.CertificateSigningRequest) (*certv1.CertificateSigningRequest, error)
}

// CreateCertificateSigningRequests creates CSR
func (c *Client) CreateCertificateSigningRequests(
	csrData []byte,
	name string,
	labels map[string]string,
	signerName string,
	requestedDuration *time.Duration,
	usages []certv1.KeyUsage,
) (*certv1.CertificateSigningRequest, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	csr := getCSR(csrData, name, labels, signerName, requestedDuration, usages)

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().Create(
		context.TODO(),
		csr,
		metav1.CreateOptions{},
	)
}

// UpdateCertificateSigningRequests updates existing CSR
func (c *Client) UpdateCertificateSigningRequests(
	csrData []byte,
	name string,
	labels map[string]string,
	signerName string,
	requestedDuration *time.Duration,
	usages []certv1.KeyUsage,
) (*certv1.CertificateSigningRequest, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	csr := getCSR(csrData, name, labels, signerName, requestedDuration, usages)

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().Update(
		context.TODO(),
		csr,
		metav1.UpdateOptions{},
	)
}

func getCSR(
	csrData []byte,
	name string,
	labels map[string]string,
	signerName string,
	requestedDuration *time.Duration,
	usages []certv1.KeyUsage,
) *certv1.CertificateSigningRequest {
	csr := &certv1.CertificateSigningRequest{
		// Username, UID, Groups will be injected by API server.
		TypeMeta: metav1.TypeMeta{
			Kind: "CertificateSigningRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request:    csrData,
			Usages:     usages,
			SignerName: signerName,
		},
	}
	if len(csr.Name) == 0 {
		csr.GenerateName = "csr-"
	}
	if requestedDuration != nil {
		i := int32(*requestedDuration / time.Second)
		csr.Spec.ExpirationSeconds = &i
	}
	return csr
}

// ListCertificateSigningRequests lists the existing certificate signing requests
func (c *Client) ListCertificateSigningRequests(labels map[string]string) (*certv1.CertificateSigningRequestList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: mapToCSV(labels),
		})
}

// GetCertificateSigningRequest retrieves a named certificate request
func (c *Client) GetCertificateSigningRequest(name string) (*certv1.CertificateSigningRequest, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().Get(
		context.TODO(),
		name,
		metav1.GetOptions{},
	)
}

// DeleteCertificateSigningRequests deletes the given CSR
func (c *Client) DeleteCertificateSigningRequests(name string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().Delete(
		context.TODO(),
		name,
		metav1.DeleteOptions{},
	)
}

// WatchCertificateSigningRequests reports changes on the requested CSR
// - CAUTION: Must populate at least csr.Name
func (c *Client) WatchCertificateSigningRequests(csr *certv1.CertificateSigningRequest, fn WatchFunc) error {
	if err := c.initClient(); err != nil {
		return err
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", csr.Name).String(),
		Watch:         true,
	}

	watchInterface, err := c.kubernetes.CertificatesV1().CertificateSigningRequests().Watch(context.TODO(), listOptions)
	if err != nil {
		return err
	}

	// fire off watch function
	go c.handleWatch(watchInterface, csr, "", fn, listOptions)
	return nil
}

// CertificateSigningRequestsUpdateApproval used to approve or decline the CSR
func (c *Client) CertificateSigningRequestsUpdateApproval(
	name string,
	csr *certv1.CertificateSigningRequest,
) (*certv1.CertificateSigningRequest, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CertificatesV1().CertificateSigningRequests().UpdateApproval(
		context.TODO(),
		name,
		csr,
		metav1.UpdateOptions{},
	)
}
