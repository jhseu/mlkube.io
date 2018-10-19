import json
import logging
from kubernetes import client as k8s_client
from kubeflow.testing import util
from py import ks_util
from py import tf_job_client

def get_runconfig(master_host, namespace, target):
  """Issue a request to get the runconfig of the specified replica running test_server.
    Args:
    master_host: The IP address of the master e.g. https://35.188.37.10
    namespace: The namespace
    target: The K8s service corresponding to the pod to call.
  """
  response = tf_operator_util.send_request(master_host, namespace, target, "runconfig", {})
  return yaml.load(response)


def verify_runconfig(master_host, namespace, job_name, replica, num_ps, num_workers):
  """Verifies that the TF RunConfig on the specified replica is the same as expected.
    Args:
    master_host: The IP address of the master e.g. https://35.188.37.10
    namespace: The namespace
    job_name: The name of the TF job
    replica: The replica type (chief, ps, or worker)
    num_ps: The number of PS replicas
    num_workers: The number of worker replicas
  """
  is_chief = True
  num_replicas = 1
  if replica == "ps":
    is_chief = False
    num_replicas = num_ps
  elif replica == "worker":
    is_chief = False
    num_replicas = num_workers

  # Construct the expected cluster spec
  chief_list = ["{name}-chief-0:2222".format(name=job_name)]
  ps_list = []
  for i in range(num_ps):
    ps_list.append("{name}-ps-{index}:2222".format(name=job_name, index=i))
  worker_list = []
  for i in range(num_workers):
    worker_list.append("{name}-worker-{index}:2222".format(name=job_name, index=i))
  cluster_spec = {
    "chief": chief_list,
    "ps": ps_list,
    "worker": worker_list,
  }

  for i in range(num_replicas):
    full_target = "{name}-{replica}-{index}".format(name=job_name, replica=replica.lower(), index=i)
    actual_config = get_runconfig(master_host, namespace, full_target)
    expected_config = {
      "task_type": replica,
      "task_id": i,
      "cluster_spec": cluster_spec,
      "is_chief": is_chief,
      "master": "grpc://{target}:2222".format(target=full_target),
      "num_worker_replicas": num_workers + 1, # Chief is also a worker
      "num_ps_replicas": num_ps,
    }
    # Compare expected and actual configs
    if actual_config != expected_config:
      msg = "Actual runconfig differs from expected. Expected: {0} Actual: {1}".format(
        str(expected_config), str(actual_config))
      logging.error(msg)
      raise RuntimeError(msg)


def run_tfjob_and_verify_runconfig(test_case, args):
  api_client = k8s_client.ApiClient()
  namespace, name, env = ks_util.setup_ks_app(args)

  # Create the TF job
  util.run(["ks", "apply", env, "-c", args.component], cwd=args.app_dir)
  logging.info("Created job %s in namespaces %s", name, namespace)

  # Wait for the job to either be in Running state or a terminal state
  logging.info("Wait for conditions Running, Succeeded, or Failed")
  results = tf_job_client.wait_for_condition(
    api_client, namespace, name, ["Running", "Succeeded", "Failed"],
    status_callback=tf_job_client.log_status)
  logging.info("Current TFJob:\n %s", json.dumps(results, indent=2))

  num_ps = results.get("spec", {}).get("tfReplicaSpecs", {}).get(
    "PS", {}).get("replicas", 0)
  num_workers = results.get("spec", {}).get("tfReplicaSpecs", {}).get(
    "Worker", {}).get("replicas", 0)
  verify_runconfig(masterHost, namespace, name, "chief", num_ps, num_workers)
  verify_runconfig(masterHost, namespace, name, "worker", num_ps, num_workers)
  verify_runconfig(masterHost, namespace, name, "ps", num_ps, num_workers)

  tf_job_client.terminate_replicas(api_client, namespace, name, "chief", 1)

  # Wait for the job to complete.
  logging.info("Waiting for job to finish.")
  results = tf_job_client.wait_for_job(
    api_client, namespace, name, args.tfjob_version,
    status_callback=tf_job_client.log_status)
  logging.info("Final TFJob:\n %s", json.dumps(results, indent=2))

  if not tf_job_client.job_succeeded(results):
    test_case.failure = "Job {0} in namespace {1} in status {2}".format(
      name, namespace, results.get("status", {}))
    logging.error(test_case.failure)
    return False

  # Delete the TFJob.
  tf_job_client.delete_tf_job(api_client, namespace, name, version=args.tfjob_version)
  logging.info("Waiting for job %s in namespaces %s to be deleted.", name,
               namespace)
  tf_job_client.wait_for_delete(
    api_client, namespace, name, args.tfjob_version, status_callback=tf_job_client.log_status)

  return True
