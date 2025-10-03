pipeline {
    agent any

    environment {
        GO111MODULE = 'on'
        NODE_OPTIONS = '--max_old_space_size=4096'
    }

    options {
        timestamps()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Backend - Test') {
            steps {
                dir('backend') {
                    sh 'go test ./...'
                    sh 'go build ./...'
                }
            }
        }

        stage('Frontend - Install Dependencies') {
            steps {
                dir('frontend') {
                    sh 'npm install'
                }
            }
        }

        stage('Frontend - Build') {
            steps {
                dir('frontend') {
                    sh 'npm run build'
                }
            }
        }

        stage('Docker - Lint Compose') {
            steps {
                sh 'docker compose config -q'
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'frontend/dist/**/*', allowEmptyArchive: true
        }
    }
}
