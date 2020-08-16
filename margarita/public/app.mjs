import "https://cdnjs.cloudflare.com/ajax/libs/mithril/2.0.4/mithril.min.js";

const Configurations = {
  list: [],
  load: () => {
    return m.request({
      method: "GET",
      url: "http://localhost:3000/api/configurations",
    }).then((result) => Configurations.list = result);
  },
  save: () => {
    return m.request({
      method: "POST",
      url: "http://localhost:3000/api/configurations",
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
    m("main", [
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
    return m("div", [
      TextField("Domain", "example.com", "domain", index, item.domain),
      TextField("Proxy", "localhost:8080", "proxy", index, item.proxy),
      m("button", {type: "button", onclick: () => Configurations.remove(index)}, "Remove"),
      m("br"), m("br"),
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
    m("button", {
      type: "button",
      onclick: () => {
        Configurations.list.push({});
      },
    }, "Add"),
    m("br")
  ]
};

const SaveButton = {
  view: () => [
    m("button", {type: "submit"}, "Save"),
    m("br")
  ]
};

m.mount(document.body, App);