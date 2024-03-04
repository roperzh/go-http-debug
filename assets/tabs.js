class DataTabsWrapper extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
  }

  connectedCallback() {
    this.renderTabs();

    const style = document.createElement("style");
    style.textContent = `
pre {
  white-space: pre-wrap;
  background-color: var(--main-surface-secondary);
  border-radius: 4px;
  padding: 10px 15px;
  margin-bottom: 46px;
}
        `;
    this.shadowRoot.appendChild(style);

    document.addEventListener("item-selected", (event) => {
      this.updateTabContents(event.detail.item);
    });
  }

  renderTabs() {
    this.shadowRoot.innerHTML = `
<tabs-container>
  <tab-item for="request">Request</tab-item>
  <tab-item for="response">Response</tab-item>
  <tab-content id="request">
    <div><pre></pre></div>
  </tab-content>
  <tab-content id="response">
    <div><pre></pre></div>
  </tab-content>
</tabs-container>
    `;
  }

  updateTabContents(itemData) {
    const requestContent = this.shadowRoot.querySelector("#request");
    const responseContent = this.shadowRoot.querySelector("#response");

    requestContent.innerHTML = `
<div>
  <h3>Headers</h3>
  <pre>${itemData.request.raw_headers.trim()}</pre>
  <h3>Body</h3>
  <pre>${itemData.request.body || "Empty request body"}</pre>
</div>
   `;
    responseContent.innerHTML = `
<div>
  <h3>Headers</h3>
  <pre>${itemData.response.raw_headers.trim()}</pre>
  
  <h3>Body</h3>
  <pre>${itemData.response.body || "Empty response body"}</pre>
</div>
   `;
  }
}

customElements.define("data-tabs-wrapper", DataTabsWrapper);

class TabsContainer extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
  }

  connectedCallback() {
    const style = document.createElement("style");
    style.textContent = `
tab-content {
  margin-bottom: 20px;
}
        `;
    this.shadowRoot.appendChild(style);

    const slot = document.createElement("slot");
    this.shadowRoot.appendChild(slot);

    slot.addEventListener("slotchange", (e) => {
      this.initTabs();
    });
  }

  initTabs() {
    const tabs = this.querySelectorAll("tab-item");
    tabs.forEach((tab, index) => {
      if (index === 0) {
        tab.classList.add("active");
        this.showTabContent(tab.getAttribute("for"));
      }

      tab.addEventListener("click", () => {
        this.resetTabs();
        tab.classList.add("active");
        this.showTabContent(tab.getAttribute("for"));
      });
    });
  }

  resetTabs() {
    const tabs = this.querySelectorAll("tab-item");
    tabs.forEach((tab) => {
      tab.classList.remove("active");
    });

    const contents = this.querySelectorAll("tab-content");
    contents.forEach((content) => {
      content.style.display = "none";
    });
  }

  showTabContent(id) {
    const content = this.querySelector(`#${id}`);
    if (content) {
      content.style.display = "block";
    }
  }
}

customElements.define("tabs-container", TabsContainer);

class TabItem extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    const style = document.createElement("style");
    style.textContent = `
:host {
  display: inline-block;
  padding: 10px;
  margin-right: 5px;
  cursor: pointer;
}
:host(.active) {
  border-bottom: 2px solid var(--brand-purple);
}
        `;
    this.shadowRoot.appendChild(style);

    const content = document.createElement("span");
    content.textContent = this.textContent;
    this.shadowRoot.appendChild(content);
  }
}

customElements.define("tab-item", TabItem);

class TabContent extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    const style = document.createElement("style");
    style.textContent = `
:host {
  display: none;
}
        `;
    this.shadowRoot.appendChild(style);

    const slot = document.createElement("slot");
    this.shadowRoot.appendChild(slot);
  }
}

customElements.define("tab-content", TabContent);
