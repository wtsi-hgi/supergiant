$(function() {

  $('[data-toggle="dropdown"]').on('click', function(event) {
    var btn = $(this),
        menu = btn.siblings('ul.dropdown-menu');
        openFn = function() {
          btn.toggleClass('open');
          menu.toggleClass('open');
        };

    openFn();

    event.stopPropagation();

    // Bind once to close out on any click
    $(document).one('click', openFn);

  });
});
