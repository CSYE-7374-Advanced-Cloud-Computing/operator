package appservice

import (
	"context"
	"encoding/json"
	"fmt"

	appv1alpha1 "github.com/CSYE-7374-Advanced-Cloud-Computing/operator/pkg/apis/app/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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

	creates3folder(cr.Spec.Username, "csye7374-operator-s3")

	user := createIamUser(cr.Spec.Username)

	// fmt.Println("generated keys: ", *user)

	accesskey := []byte(*user.AccessKeyId)
	secretkey := []byte(*user.SecretAccessKey)

	cr.Status.Setupcomplete = true

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Secretname,
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			"AccesKey":  accesskey,
			"SecretKey": secretkey,
		},
	}
}

func creates3folder(folderName string, bucket string) {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Profile: "kops",
	})

	svc := s3.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	params := s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &folderName,
	}

	_, err := svc.PutObject(&params)

	fmt.Println("Folder ", folderName, "created in ", bucket, "bucket")

	if err != nil {
		fmt.Println("S3 Error: ", err)
	}
}

func createIamUser(username string) *iam.AccessKey {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Profile: "kops",
	})
	svc := iam.New(sess)

	user, err := svc.GetUser(&iam.GetUserInput{
		UserName: &username,
	})

	if err != nil {
		if awserr, ok := err.(awserr.Error); ok && awserr.Code() == "NoSuchEntity" {
			result, err := svc.CreateUser(&iam.CreateUserInput{
				UserName: &username,
			})
			if err != nil {
				fmt.Println("create user error:", err)
			} else {
				fmt.Println("New User Created: \n", result.User)
				key := createAccessKey(*result.User.UserName)
				createPolicy(username, "csye7374-operator-s3", "295717451775")
				return key
			}
		}
	}
	fmt.Println("User present: ", user.User)
	return createAccessKey(*user.User.UserName)
}

func createAccessKey(username string) *iam.AccessKey {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Profile: "kops",
	})
	svc := iam.New(sess)

	keyInput := &iam.ListAccessKeysInput{
		UserName: &username,
	}

	key, _ := svc.ListAccessKeys(keyInput)

	// fmt.Println(*key)
	// fmt.Println(err)

	input := &iam.CreateAccessKeyInput{
		UserName: &username,
	}

	if len(*&key.AccessKeyMetadata) == 0 {
		result, _ := svc.CreateAccessKey(input)
		fmt.Println("User Credentials Created: \n", result.AccessKey)
		// fmt.Println(result.AccessKey)
		// fmt.Println(err)
		return result.AccessKey
	} else {
		keyDeleteInput := iam.DeleteAccessKeyInput{
			AccessKeyId: *&key.AccessKeyMetadata[0].AccessKeyId,
			UserName:    &username,
		}
		svc.DeleteAccessKey(&keyDeleteInput)
		fmt.Println("Credentials Deleted")
		result, _ := svc.CreateAccessKey(input)
		fmt.Println("User Credentials Created: \n", result.AccessKey)
		// fmt.Println(reflect.TypeOf(result.AccessKey))
		// fmt.Println(err)
		return result.AccessKey
	}
}

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect   string
	Action   []string
	Resource *string
}

func createPolicy(username string, bucket string, awsaccount string) {
	sess, _ := session.NewSessionWithOptions(session.Options{
		Profile: "kops",
	})
	svc := iam.New(sess)

	policy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []StatementEntry{
			StatementEntry{
				Effect: "Allow",
				Action: []string{
					"s3:*",
				},
				Resource: aws.String("arn:aws:s3:::" + bucket + "/" + username + "/*"),
			},
		},
	}

	b, err := json.Marshal(&policy)
	if err != nil {
		fmt.Println(err)
	}
	policyinput := iam.CreatePolicyInput{
		PolicyDocument: aws.String(string(b)),
		PolicyName:     aws.String("s3-" + username),
	}

	arn := aws.String("arn:aws:iam::" + awsaccount + ":policy/s3-" + username)

	getpolicyinput := iam.GetPolicyInput{
		PolicyArn: arn,
	}

	_, policyerr := svc.GetPolicy(&getpolicyinput)

	if policyerr != nil {
		awserr, ok := policyerr.(awserr.Error)
		if ok && awserr.Code() == "NoSuchEntity" {
			fmt.Println("Creating a user Policy")
			result, _ := svc.CreatePolicy(&policyinput)
			fmt.Println("User Policy Created:", *result)
		}
	}
	attachinput := iam.AttachUserPolicyInput{
		PolicyArn: arn,
		UserName:  &username,
	}

	svc.AttachUserPolicy(&attachinput)
	fmt.Println("User Policy attached to user ", username)

}
