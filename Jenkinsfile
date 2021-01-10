pipeline {
    agent any
    triggers {
        cron('H 2 * * *')
    }
    agent {
        dockerfile { filename 'Dockerfile.build' }
    }
    environment {
        NAMESPACE = 'registry'
        CONFIG = 'bin/clean/config.yaml'
        REGISTRY_URL = 'http://localhost:5000'
        REGISTRY_USER = credentials('jenkins-registry-user')
        REGISTRY_PASSWORD = credentials('jenkins-registry-password')
    }
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Soft-delete manifests') {
            steps {
                script {
                    sh 'python -m venv /tmp/venv-registry-clean'
                    sh 'source /tmp/venv-registry-clean/bin/activate'
                    sh 'pip install -r bin/clean/requirements.txt'
                    sh 'python bin/clean/main.py'
                }
            }
        }
        stage('Run garbage collector') {
            steps {
                script {
                    sh './bin/maintenance.sh on'
                    sh './bin/garbage-collector.sh'
                }
            }
        }
    }
    post {
        failure {
            msg = "Error: failed to clean registry."
            slackSend message: msg, channel: env.SLACK_CHANNEL
        }
        always {
            script {
                sh './bin/maintenance.sh off'
                sh 'rm -rf /tmp/venv-registry-clean'
            }
        }
    }
}
