pipeline {
    options {
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '5'))
    }
    triggers {
        cron('@monthly')
    }
    agent {
        docker {
            image 'golang'
            label "docker"
        }
    }
    stages {
        stage("Prepare dependencies") {
            steps {
                sh 'curl -Ls https://github.com/goreleaser/goreleaser/releases/download/v0.82.2/goreleaser_Linux_x86_64.tar.gz | tar -zxv goreleaser'
                sh 'go get -v -t -d'
            }
        }
        stage("Test build") {
            when {
	              not { tag "v*" }
            }
            steps {
                sh './goreleaser release --snapshot --rm-dist'
            }
        }
        stage("goreleaser tagged release") {
            when {
	              tag "v*"
            }
            steps {
                sh './goreleaser release --rm-dist'
            }
        }
    }
}
