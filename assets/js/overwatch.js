var Overwatch = {};
Overwatch['Model'] = {};
Overwatch['View']  = {};
Overwatch['Collection'] = {};



Overwatch.Model.LogLine = Backbone.Model.extend({
  defaults: {
    "channel"   : "NA",
    "content"   : "NA"
  }
});



Overwatch.Collection.LogLines = Backbone.Collection.extend({
  counter:   0,
  maxSize: 100,
  model: Overwatch.Model.LogLine,

  initialize: function() {
    this.listenTo(this, "add", function(e) {
      this.counter += 1;

      // if we have too many elements, remove the first one!
      if (this.length > this.maxSize) {
        this.shift();
      }
    });
  }

});



Overwatch.View.Navigation = Backbone.View.extend({
  events: {
    "keyup .search"  : "search",
    "change .search" : "search",
    "submit form"    : "submit"
  },

  initialize: function() {
    this.searchInput = this.$(".search");
    this.inSearch    = false;
  },

  submit: function(e) {
    e.preventDefault();
    return
  },

  search: function(evt) {
    var query = this.searchInput.val();
    var that  = this;

    /* gross but simple part 1 */
    if (query.length <= 2 && this.inSearch == true) {
      $('.terminal-row').each(function(i,e){
        $(e).removeAttr('style');
      });
      this.inSearch = false;
      return
    }

    $('.terminal-row').each(function(i,e){
      if (! e.innerText.match(query)) {
        $(e).css('opacity', '0.4')
        that.inSearch = true;
      } else {
        $(e).removeAttr('style');
        that.inSearch = true;
      }
    });
  }

});



Overwatch.View.Terminal = Backbone.View.extend({

  lineTemplate : function(){
    var tmpl = '<div class="terminal-row">';
    tmpl += '<span class="terminal-text-count">{{count}}</span>';
    tmpl += '<span class="terminal-text-pipe">&nbsp;|&nbsp;</span>';
    tmpl += '<span class="terminal-text-channel">{{channel}}</span>';
    tmpl += '<span class="terminal-text-pipe">&nbsp;|&nbsp;</span>';
    tmpl += '<span class="terminal-text-line">{{content}}</span>';
    tmpl += "</div>"
    return tmpl;
  }(),

  initialize: function() {
    this.listenTo(this.model, "add",    this.appendLine);
    this.listenTo(this.model, "remove", this.removeLine);
  },

  appendLine: function(e) {
    var view = {
      count   : this.model.counter,
      channel : e.get('channel'),
      content : ansi_up.ansi_to_html(e.get('content'))
    };

    // make lines a little more pretty
    // regex_date1 = /(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Dec)\s(\d{1,2})\s(\d{2}\:\d{2}\:\d{2})/;
    // if (blah.match(regex_date1)) {
    // line = line.replace(regex_date1, '<span class="terminal-text-date">$1 $2 $3</span>');
    // }

    this.$el.append(Mustache.render(this.lineTemplate, view));
    this.$el.parent().scrollTop(this.$el[0].scrollHeight);

    /* gross but simple part 2*/
    var query = $('.search').val();
    if ( query.length > 2 ){
      var last = $('.terminal-row:last')[0]
      if (! last.innerText.match(query) ) {
        $(last).css('opacity', 0.4);
      }
    }
  },

  removeLine: function() {
    $(this.$el.children()[0]).remove();
  }
});
