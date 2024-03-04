class StatusCode extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
  }

  connectedCallback() {
    const style = document.createElement("style");
    style.textContent = `
             :host {
	     padding: 2px 6px;
  border-radius: 6px;
  font-weight: bold;
      	    }

	    :host(.green) {
		background-color: green;
	    }

	    :host(.yellow) {
	    	background-color: yellow;
	    }

	    :host(.orange) {
	    	background-color: orange;
	    }

	    :host(.red) {
	    	background-color: red;
	    }
              `;
    this.shadowRoot.appendChild(style);

    const code = parseInt(this.textContent, 10);
    var klass = "";
    if (code < 299) {
      klass = "green";
    } else if (code <= 399) {
      klass = "yellow";
    } else if (code <= 499) {
      klass = "orange";
    } else {
      klass = "red";
    }

    this.classList.add(klass);
    const content = document.createElement("span");
    content.textContent = this.textContent;
    this.shadowRoot.appendChild(content);
  }
}
customElements.define("status-code", StatusCode);

class ResponseItem extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
  }

  connectedCallback() {
    const style = document.createElement("style");
    style.textContent = `
            div {
              padding: 16px;
              cursor: pointer;
            }

            div:hover {
              background-color: var(--main-surface-tertiary);
            }

            :host(.active) div {
              background-color: var(--main-surface-secondary);
            }

	    status-code {
	    margin-right: 8px;
	    }
                    `;
    this.shadowRoot.appendChild(style);
  }

  set item(data) {
    const div = document.createElement("div");
    div.innerHTML = `<status-code>${data.status}</status-code> <strong>${data.method}</strong> <i>${data.path}</i>`;
    this.shadowRoot.appendChild(div);
  }
}

customElements.define("response-item", ResponseItem);

class ResultsList extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
  }

  async connectedCallback() {
    const response = await fetch("/data");
    const data = await response.json();

    const list = document.createElement("div");
    this.shadowRoot.appendChild(list);

    const responseItems = [];

    if (Array.isArray(data) && data.length) {
      data.forEach((item, index) => {
        const responseItem = document.createElement("response-item");
        responseItems.push(responseItem);
        responseItem.item = item;
        responseItem.addEventListener("click", () => {
          responseItems.forEach((ri) => ri.classList.remove("active"));
          responseItem.classList.add("active");
          this.dispatchEvent(
            new CustomEvent("item-selected", {
              bubbles: true,
              detail: { item },
            }),
          );
        });

        this.shadowRoot.appendChild(responseItem);

        if (index === 0) {
          setTimeout(() => responseItem.click(), 100);
        }
      });
    } else {
      list.textContent = "No results found.";
    }
  }
}

customElements.define("results-list", ResultsList);
