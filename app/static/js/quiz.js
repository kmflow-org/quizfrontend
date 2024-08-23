// static/js/quiz.js

document.addEventListener("DOMContentLoaded", function () {
    document.getElementById("submitQuiz").addEventListener("click", function () {
        submitQuiz();
    });

    function submitQuiz() {
        let form = document.getElementById('quizForm');
        let formData = new FormData(form);
        let quizData = {};

        formData.forEach(function(value, key) {
            if (!quizData[key]) {
                quizData[key] = [];
            }
            quizData[key].push(value);
        });

        // Convert the quizData object to JSON.
        let jsonData = JSON.stringify(quizData);

        // Make the AJAX request.
        fetch('/evaluate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: jsonData
        })
        .then(response => response.json())
        .then(data => {
            let results = document.getElementById("results");
            results.innerHTML="";
            data.forEach((result)=>{
                results.innerHTML+="<h3>Question: "+result["questionId"] + " correct: "+ result["correct"]+"</h3><br/>";
            });
        })
        .catch((error) => {
            console.error('Error:', error);
            alert('An error occurred while submitting the quiz.');
        });
    }
});
