pipeline {
    agent any

    environment {
        // Ganti 'namakampusacr' dengan nama Azure Container Registry (ACR) Anda
        ACR_SERVER    = 'namakampusacr.azurecr.io'
        HUB_IMAGE     = "${ACR_SERVER}/logistics-hub:latest"
        COURIER_IMAGE = "${ACR_SERVER}/logistics-courier:latest"
        
        // Resource Group & Cluster Name untuk AKS (Azure Kubernetes Service)
        AKS_RESOURCE_GROUP = 'logistics-rg'
        AKS_CLUSTER_NAME   = 'logistics-aks'
    }

    stages {
        stage('1. Checkout Repo') {
            steps {
                echo 'Checking out source code...'
                checkout scm
            }
        }

        stage('2. Unit Tests') {
            steps {
                echo 'Running Unit Tests...'
                // Pastikan Golang sudah terinstall di agent Jenkins
                sh 'go test -v ./internal/...'
            }
        }

        stage('3. Lint / Vet') {
            steps {
                echo 'Running Linter and Vet...'
                sh 'go vet ./...'
            }
        }

        stage('4. Build Image (lokal)') {
            steps {
                echo 'Building Docker Images...'
                sh "docker build -f deployments/docker/hub.Dockerfile -t ${HUB_IMAGE} ."
                sh "docker build -f deployments/docker/courier.Dockerfile -t ${COURIER_IMAGE} ."
            }
        }

        stage('5. Functional Tests') {
            steps {
                echo 'Starting local environment with Docker Compose for Functional Test...'
                // Menyalakan DB, App, dan Mock server
                sh 'docker compose -f deployments/docker/docker-compose.yml up -d'
                
                // Menunggu service ready (healthcheck sederhana)
                sleep 10
                
                // Eksekusi script functional test
                // (Untuk tugas ini, ini merepresentasikan aksi curl/postman otomatis)
                echo 'Running E2E Functional API Tests...'
                sh 'go test -v ./tests/functional/...'
                
                // Matikan environment setelah selesai test
                sh 'docker compose -f deployments/docker/docker-compose.yml down'
            }
        }

        stage('6. Push image (ke Azure Container Registry)') {
            steps {
                echo 'Pushing Docker Images to Azure Container Registry (ACR)...'
                // Login ke ACR (Bisa diotomasi dengan Jenkins Credentials / Service Principal)
                // sh 'az acr login --name namakampusacr'
                
                // Push image ke cloud Azure
                // sh "docker push ${HUB_IMAGE}"
                // sh "docker push ${COURIER_IMAGE}"
                echo "Simulating push for ${HUB_IMAGE} and ${COURIER_IMAGE} to ACR"
            }
        }

        stage('7. Deploy di Azure Kubernetes Service (AKS)') {
            steps {
                echo 'Deploying to Azure Kubernetes Cluster...'
                // Mendapatkan credentials K8s dari Azure sebelum deploy
                // sh "az aks get-credentials --resource-group ${AKS_RESOURCE_GROUP} --name ${AKS_CLUSTER_NAME}"
                
                // Apply file YAML ke Azure K8s
                sh 'kubectl apply -f deployments/k8s/'
            }
        }

        stage('8. Verify') {
            steps {
                echo 'Verifying Kubernetes Deployment...'
                sh 'kubectl get pods'
                sh 'kubectl get services'
                
                // Pastikan status deployment berhasil rollout
                sh 'kubectl rollout status deployment/hub-deployment'
                sh 'kubectl rollout status deployment/courier-deployment'
            }
        }
    }

    post {
        always {
            echo 'Pipeline execution complete.'
        }
        success {
            echo 'All stages passed! The Logistics System is Production-Ready. 🚀'
        }
        failure {
            echo 'Pipeline failed. Please check the logs.'
            // Menjamin docker-compose mati walau test gagal
            sh 'docker compose -f deployments/docker/docker-compose.yml down || true'
        }
    }
}
