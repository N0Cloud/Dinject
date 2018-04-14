pipeline {
  agent any
  stages {
    stage('prepare') {
      steps {
        node(label: 'docker') {
          sh 'go version'
        }

      }
    }
  }
}