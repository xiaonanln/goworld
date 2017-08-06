cc.Class({
    extends: cc.Component,

    properties: {
        // foo: {
        //    default: null,      // The default value will be used only when the component attaching
        //                           to a node for the first time
        //    url: cc.Texture2D,  // optional, default is typeof default
        //    serializable: true, // optional, default is true
        //    visible: true,      // optional, default is true
        //    displayName: 'Foo', // optional
        //    readonly: false,    // optional, default is false
        // },
        // ...
        typeName: {
            default: '',  
            readonly: true, 
        },
        
        entityID: {
            default: '', 
            readonly: true, 
        }
        
    },

    // use this for initialization
    onLoad: function () {

    },

    create: function(typeName, entityID) {
        this.typeName = typeName
        this.entityID = entityID
    },

    toString: function() {
        return this.typeName + "<" + this.entityID + ">"
    },

    onCreated: function() {
        if (this.typeName == "Account") {
            console.log("Account created, start logining...")
        }
    },

    // called every frame, uncomment this function to activate update callback
    // update: function (dt) {

    // },
});
