// static/js/quiz.js

document.addEventListener("DOMContentLoaded", function () {
    document.getElementById("submitQuiz").addEventListener("click", function () {
        submitQuiz();
    });

    function clearQuestionResults() {
        document.querySelectorAll(".question").forEach((questionEl) => {
            questionEl.classList.remove("is-correct", "is-incorrect");
        });

        document.querySelectorAll(".question-result").forEach((resultEl) => {
            resultEl.classList.remove("visible", "correct", "incorrect");
            resultEl.textContent = "";
        });
    }

    function renderResultForQuestion(result) {
        const questionId = String(result["questionId"]);
        const isCorrect = Boolean(result["correct"]);

        const resultEl = document.getElementById(`result-${questionId}`);
        const questionCardEl = document.querySelector(`.question[data-question-id="${questionId}"]`);

        if (!resultEl || !questionCardEl) {
            return;
        }

        resultEl.classList.add("visible", isCorrect ? "correct" : "incorrect");
        resultEl.textContent = isCorrect ? "Correct answer" : "Incorrect answer";

        questionCardEl.classList.add(isCorrect ? "is-correct" : "is-incorrect");
    }

    function submitQuiz() {
        const form = document.getElementById('quizForm');
        const formData = new FormData(form);
        const quizData = {};

        formData.forEach(function(value, key) {
            if (!quizData[key]) {
                quizData[key] = [];
            }
            quizData[key].push(value);
        });

        // Convert the quizData object to JSON.
        const jsonData = JSON.stringify(quizData);

        clearQuestionResults();

        // Make the AJAX request.
        fetch('/evaluate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: jsonData
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Failed with status ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            data.forEach((result) => {
                renderResultForQuestion(result);
            });
        })
        .catch((error) => {
            console.error('Error:', error);
            alert('An error occurred while submitting the quiz.');
        });
    }
});
