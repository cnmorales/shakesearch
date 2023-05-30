const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}&ww=${data.ww}&cs=${data.cs}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const thead = document.getElementById("table-head");
    thead.innerHTML = "<tr><th>Results</th></tr>"

    const table = document.getElementById("table-body");
    const rows = [];

    for (let result of results) {
      result = result.replace(/\r\n|\n|\r/gm, '<br />')
      rows.push(`<tr><th>${result}<tr/></th>`);
    }
    table.innerHTML = rows;
  },

};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);

// getting all required elements
const searchInput = document.querySelector(".searchInput");
const input = searchInput.querySelector("input");
const resultBox = searchInput.querySelector(".resultBox");
const icon = searchInput.querySelector(".icon");
let linkTag = searchInput.querySelector("a");
let webLink;

function showResults(val) {
  if (val.length < 3) {
      searchInput.classList.remove("active"); //hide autocomplete box
      return;
  }
  res = document.getElementById("result");
  res.innerHTML = '';
  if (val == '') {
    return;
  }
  let list = '';
  fetch('/suggest?q=' + val).then(
      function (response) {
        return response.json();
      }).then(function (data) {
    for (i=0; i<data.length; i++) {
      list += '<li>' + data[i] + '</li>';
    }
    res.innerHTML = '<ul>' + list + '</ul>';

    document.addEventListener('click', clickOutside);
    searchInput.classList.add("active"); //show autocomplete box

    let allList = resultBox.querySelectorAll("li");
    for (let i = 0; i < allList.length; i++) {
      //adding onclick attribute in all li tag
      //allList[i].setAttribute("onclick", "select(this)");
      allList[i].addEventListener('click', function () {
        input.value = this.innerHTML;
        closeList();
      });
    }

    return true;
  }).catch(function (err) {
    console.warn('Something went wrong.', err);
    return false;
  });
}

function closeList() {
  searchInput.classList.remove("active");
  document.removeEventListener('click', clickOutside);
}

function clickOutside(event) {
  if (event.target !== input) {
    document.removeEventListener('click', clickOutside);
    searchInput.classList.remove("active");
  }
}