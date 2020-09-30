import "https://cdnjs.cloudflare.com/ajax/libs/mithril/2.0.4/mithril.min.js";

const Configurations = {
  list: [],
  load: () => {
    return m.request({
      method: "GET",
      url: "http://localhost:10010/api/configurations",
    }).then((result) => Configurations.list = result);
  },
  save: () => {
    return m.request({
      method: "POST",
      url: "http://localhost:10010/api/configurations",
      body: Configurations.list,
    }).then(() => console.log("saved"));
  },
  update: (index, field, value) => {
    Configurations.list[index][field] = value;
  },
  remove: (index) => {
    Configurations.list.splice(index, 1);
  },
};

const App = {
  view: () => [
    m("main", {class: "container"}, [
      m("h1", "Poolside Development"),
      m("hr"),
      m("h3", "Configurations"),
      m(Form),
    ])
  ]
};

const Form = {
  onload: Configurations.load(),
  view: () => [
    m("form", {
      onsubmit: (e) => {
        e.preventDefault();
        Configurations.save();
      }
    }, [
      TextFieldList(Configurations.list),
      m(AddButton),
      m(SaveButton),
    ])
  ]
}

const TextFieldList = (list) =>
  list.map((item, index) => {
    return m("div", {class: "row"}, [
      m("div", {class: "column"}, [
        TextField("Domain", "example.com", "domain", index, item.domain),
      ]),
      m("div", {class: "column"}, [
        TextField("Proxy", "localhost:8080", "proxy", index, item.proxy),
      ]),
      m("div", {class: "column column-10", style: "margin: auto;text-align: center;padding-top:25px;"}, [
        m("button", {type: "button", onclick: () => Configurations.remove(index)}, "Remove"),
      ]),
    ]);
  });

const TextField = (label, placeholder, field, index, value) =>
  m("div", [
    m("label", {for: field}, `${label}: `),
    m("input", {
      type: "text",
      id: field,
      name: field,
      placeholder: placeholder,
      value: value,
      oninput: (e) => Configurations.update(index, field, e.target.value),
    }),
    m("br"),
  ]);

const AddButton = {
  view: () => [
    m("div", {style: "text-align: center;"}, [
      m("button", {
        type: "button",
        class: "button-outline",
        onclick: () => {
          Configurations.list.push({});
        },
      }, "Add"),
      m("br"),
    ])
  ]
};

const SaveButton = {
  view: () => [
    m("button", {type: "submit"}, "Save"),
    m("br")
  ]
};

m.mount(document.body, App);