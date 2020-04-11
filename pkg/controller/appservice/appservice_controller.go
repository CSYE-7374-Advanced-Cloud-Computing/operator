package appservice

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/CSYE-7374-Advanced-Cloud-Computing/operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	// import for aws sdks
)

type AwsAccessKey struct {
	AccessKeyId     int
	SecretAccessKey string
	UserName        string
}

type AwsCredentials struct {
	AccessKey AwsAccessKey
}

var log = logf.Log.WithName("controller_appservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAppService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppService
	err = c.Watch(&source.Kind{Type: &appv1alpha1.AppService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AppService
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AppService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAppService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppService{}

// ReconcileAppService reconciles a AppService object
type ReconcileAppService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AppService object and makes changes based on the state read
// and what is in the AppService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppService")

	// Fetch the AppService instance
	instance := &appv1alpha1.AppService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	// pod := newPodForCR(instance)
	user_secret := newSecretForCR(instance)

	// Set AppService instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, user_secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: user_secret.Name, Namespace: user_secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new secret", "secret.Namespace", user_secret.Namespace, "secret.Name", user_secret.Name)
		err = r.client.Create(context.TODO(), user_secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: secret already exists", "secret.Namespace", found.Namespace, "secret.Name", found.Name)
	return reconcile.Result{}, nil
}

// PolicyDocument is our definition of our policies to be uploaded to IAM.
type PolicyDocument struct {
	Version   string
	Statement []PolicyStatementEntry
}

// PolicyStatementEntry will dictate what this policy will allow or not allow.
type PolicyStatementEntry struct {
	Effect   string
	Action   []string
	Resource string
}

// secrets implements SecretInterface
type secrets struct {
	client rest.Interface
	ns     string
}

// Get takes name of the secret, and returns the corresponding secret object, and an error if there is any.
func (c *secrets) GetSecrets(name string, options metav1.GetOptions) (result *v1.Secret, err error) {
	result = &v1.Secret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("secrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// function to create IAM policy
func createIAMPolicy(username *string, AWS_ACCESS_KEY *string, AWS_SECRET_KEY *string, bucket *string) {
	fmt.Println("Function to create IAM Policy")
}

// function to create IAM user and attach policy
func createIAMUser(username *string, policyname *string, AWS_ACCESS_KEY_ID_1 *string, AWS_SECRETE_KEY_ID_1 *string) int {
	fmt.Println("Function to create IAM User")
	return 0
}

// function to check if the S3 bucket exists or not
func check_if_s3_bucket_exist(username string, AWS_ACCESS_KEY_ID string, AWS_SECRETE_KEY_ID string) bool {
	if username != "" {
		return true
	}
	return false
}

// function to generate Access key
func generateAccessKey(username *string, AWS_ACCESS_KEY_ID_1 *string, AWS_SECRETE_KEY_ID_1 *string, kubeClient client.Client, namespace string) {
	fmt.Println("Function to generate access key")
}

// function to create s3 bucket and put object
func createS3Bucket(username *string, AWS_ACCESS_KEY_ID *string, AWS_SECRETE_KEY_ID *string, bucket *string) {
	fmt.Println("Function to create s3 bucket and add object")
}

// GetSecret returns a secret based on a secretName and namespace.
func GetSecret(kubeClient client.Client, secretName, namespace string) (*corev1.Secret, error) {

	s := &corev1.Secret{}
	fmt.Println(" ====-  In GetSecret function -=====")

	err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: namespace}, s)

	if err != nil {
		return nil, err
	}

	return s, nil
}

// Function implementing main logic
func coreOperations(r *ReconcileAppService, cr *appv1alpha1.AppService) {

	fmt.Println("coreOperations Function =========- \n ")

	crdSecret, err := GetSecret(r.client, cr.Spec.Secretname, cr.Namespace)
	fmt.Println(cr.Spec.Username)
	fmt.Println("User Secret : ", cr.Spec.Secretname)
	fmt.Println("Show current cr.Status.Setupcomplete")
	fmt.Println(cr.Status.Setupcomplete)

	if crdSecret != nil {
		fmt.Println("Get the access key, secret and bucket")
	}

	if err != nil {
		fmt.Println("Error Encountered")
	}

	if crdSecret != nil {
		bucketName := string(crdSecret.Data["s3_bucker_name"])
		aws_access_key_id := string(crdSecret.Data["aws_access_key_id"])
		aws_secret_access_key := string(crdSecret.Data["aws_secret_access_key"])

		fmt.Println(bucketName)
		fmt.Println(aws_access_key_id)
		fmt.Println(aws_secret_access_key)

		fmt.Println("Call createIAMPolicy function -- Line 181")
		fmt.Println("Call createIAMUser function -- Line 186")
		fmt.Println("Call createS3Bucket function -- Line 205")
		fmt.Println("Call the generate access key function -- 200")

		fmt.Println("Once the folder, IAM User, IAM Policy and the poliocy attached to IAM, set the setupcomplete field status to True")
		cr.Status.Setupcomplete = true
		fmt.Println(cr.Status.Setupcomplete)
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *appv1alpha1.AppService) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

func newSecretForCR(cr *appv1alpha1.AppService) *corev1.Secret {
	b := []byte("hemal")
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Secretname,
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			"AccesKey": b,
		},
	}
}

func createOperations(r *ReconcileAppService, cr *appv1alpha1.AppService) {

}
